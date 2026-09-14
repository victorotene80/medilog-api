package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"go.uber.org/zap"
)

func TestRateLimiter_InMemoryFallback_RedisDown(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_ip",
		Limit:     2,
		Window:    time.Minute,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusOK, w2.Code)

	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

func TestRateLimiter_InMemory_WindowReset(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_window",
		Limit:     1,
		Window:    100 * time.Millisecond,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	time.Sleep(150 * time.Millisecond)

	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRateLimiter_InMemory_Eviction(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_evict",
		Limit:     1,
		Window:    50 * time.Millisecond,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.2:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	time.Sleep(100 * time.Millisecond)

	rl.evictExpired()

	rl.mu.RLock()
	count := len(rl.counters)
	rl.mu.RUnlock()
	assert.Equal(t, 0, count)
}

func TestRateLimiter_Redis_AllowWithinLimit(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(rc, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_redis",
		Limit:     3,
		Window:    time.Minute,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should pass", i+1)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiter_Redis_WindowReset(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(rc, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_redis_reset",
		Limit:     1,
		Window:    2 * time.Second,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.3:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	mr.FastForward(3 * time.Second)

	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRateLimiter_Redis_FallbackOnRedisError(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(rc, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_fallback",
		Limit:     2,
		Window:    time.Minute,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.4:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	mr.Close()

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusOK, w2.Code)

	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req)
	assert.Equal(t, http.StatusOK, w3.Code)

	w4 := httptest.NewRecorder()
	handler.ServeHTTP(w4, req)
	assert.Equal(t, http.StatusTooManyRequests, w4.Code)
}

func TestRateLimiter_NilRedis_Noop(t *testing.T) {
	rl := &RateLimiter{redis: nil, counters: make(map[string]*inMemoryCounter)}

	var called bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Limit(RateLimitRule{
		Name:      "test_nil",
		Limit:     1,
		Window:    time.Minute,
		KeyFields: []string{"ip"},
	})(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_EmptyRules_Noop(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(rc, zap.NewNop())

	var called bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Limit()(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_DifferentKeys_IndependentLimits(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_independent",
		Limit:     1,
		Window:    time.Minute,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "10.0.0.10:12345"

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.11:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req1)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

func TestRateLimiter_RateLimitHeaders(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_headers",
		Limit:     5,
		Window:    time.Minute,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.20:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, "5", w1.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "4", w1.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimiter_BodyFields_RateLimit(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:       "test_body",
		Limit:      1,
		Window:     time.Minute,
		KeyFields:  []string{"ip"},
		BodyFields: []string{"email"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := []byte(`{"email":"test@example.com"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req1.RemoteAddr = "10.0.0.30:12345"
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	body2 := []byte(`{"email":"test@example.com"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body2))
	req2.RemoteAddr = "10.0.0.30:12345"
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimiter_DeviceFingerprint_Key(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "test_device",
		Limit:     1,
		Window:    time.Minute,
		KeyFields: []string{"device"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.40:12345"
	req.Header.Set("X-Device-Fingerprint", "device-abc-123")

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimiter_BodyTooLarge(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:       "test_large_body",
		Limit:      10,
		Window:     time.Minute,
		KeyFields:  []string{"ip"},
		BodyFields: []string{"email"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	largeBody := make([]byte, maxRateLimitBodySize+1)
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(largeBody))
	req.RemoteAddr = "10.0.0.50:12345"
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestRateLimiter_InvalidJSON_BodyIgnored(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:       "test_invalid_json",
		Limit:      1,
		Window:     time.Minute,
		KeyFields:  []string{"ip"},
		BodyFields: []string{"email"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", io.NopCloser(bytes.NewReader([]byte("not json"))))
	req.RemoteAddr = "10.0.0.60:12345"
	req.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)
}

// A key field that rawKey does not understand contributes nothing to the key,
// which silently merges every caller into one bucket. That is how "device"
// survived in three rules after its arm was removed, so the guard has to reject
// it at wiring time rather than trust the declaration.
func TestValidateRules_RejectsUnknownKeyField(t *testing.T) {
	err := ValidateRules(RateLimitRule{
		Name:      "auth_login_identity",
		Limit:     10,
		Window:    time.Hour,
		KeyFields: []string{"ip", "device"},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "device")
	assert.Contains(t, err.Error(), "auth_login_identity")
}

func TestValidateRules_AcceptsKnownKeyFields(t *testing.T) {
	assert.NoError(t, ValidateRules(
		RateLimitRule{Name: "a", Limit: 300, Window: time.Minute, KeyFields: []string{"user"}},
		RateLimitRule{Name: "b", Limit: 10, Window: time.Hour, KeyFields: []string{"ip", "path"}},
	))
}

func TestValidateRules_RejectsNonPositiveLimit(t *testing.T) {
	err := ValidateRules(RateLimitRule{Name: "c", Limit: 0, Window: time.Minute})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
}

// The per-user bucket must actually separate users, or the group-level default
// is one shared allowance for the whole authenticated surface.
func TestRateLimiter_UserKey_SeparatesUsers(t *testing.T) {
	rl := NewRateLimiter(nil, zap.NewNop())
	rule := RateLimitRule{
		Name:      "authenticated_default_user",
		Limit:     1,
		Window:    time.Minute,
		KeyFields: []string{"user"},
	}

	req := func(userID string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		return r.WithContext(context.WithValue(
			r.Context(),
			appContracts.AuthContextKey,
			appContracts.AuthContext{UserID: userID},
		))
	}

	keyA := rl.rawKey(req("101"), rule, nil)
	keyB := rl.rawKey(req("202"), rule, nil)
	keyAAgain := rl.rawKey(req("101"), rule, nil)

	assert.NotEqual(t, keyA, keyB, "two users must not share a rate-limit bucket")
	assert.Equal(t, keyA, keyAAgain, "the same user must map to a stable bucket")
}

// The window is enforced entirely by the key's TTL. INCR on a missing key
// creates it without one, so the expiry must be applied in the same atomic step
// — otherwise a failure between the two commands strands the key at TTL -1 and
// it increments forever.
func TestRateLimiter_Redis_CounterAlwaysCarriesTTL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(rc, zap.NewNop())

	handler := rl.Limit(RateLimitRule{
		Name:      "ttl_probe",
		Limit:     5,
		Window:    time.Minute,
		KeyFields: []string{"ip"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.1.2.3:5555"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	keys := mr.Keys()
	require.Len(t, keys, 1)
	assert.Greater(t, mr.TTL(keys[0]), time.Duration(0),
		"counter key must have an expiry or the window never resets")
}

// A key left without an expiry by the old non-atomic path would 429 that caller
// permanently. The script re-applies the TTL whenever it is missing, so such a
// key heals on its next request instead of needing a manual DEL.
func TestRateLimiter_Redis_HealsKeyStrandedWithoutTTL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(rc, zap.NewNop())

	rule := RateLimitRule{
		Name:      "stranded",
		Limit:     5,
		Window:    time.Minute,
		KeyFields: []string{"ip"},
	}
	handler := rl.Limit(rule)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.9.9.9:4444"

	// Establish the key, then strip its expiry to reproduce the stranded state.
	handler.ServeHTTP(httptest.NewRecorder(), req)
	keys := mr.Keys()
	require.Len(t, keys, 1)
	require.NoError(t, rc.Persist(context.Background(), keys[0]).Err())
	require.Equal(t, time.Duration(0), mr.TTL(keys[0]))

	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.Greater(t, mr.TTL(keys[0]), time.Duration(0),
		"a key found without a TTL must have one re-applied")
}
