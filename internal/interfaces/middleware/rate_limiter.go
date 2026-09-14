package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
	"go.uber.org/zap"
)

const maxRateLimitBodySize = 1 << 20

type RateLimitRule struct {
	Name       string
	Limit      int64
	Window     time.Duration
	KeyFields  []string
	BodyFields []string
}

type inMemoryCounter struct {
	count     int64
	expiresAt time.Time
}

type RateLimiter struct {
	redis    *redis.Client
	logger   *zap.Logger
	prefix   string
	mu       sync.RWMutex
	counters map[string]*inMemoryCounter
	done     chan struct{}
}

func NewRateLimiter(redisClient *redis.Client, logger *zap.Logger) *RateLimiter {
	if logger == nil {
		logger = zap.NewNop()
	}

	rl := &RateLimiter{
		redis:    redisClient,
		logger:   logger,
		prefix:   "rate_limit:",
		counters: make(map[string]*inMemoryCounter),
		done:     make(chan struct{}),
	}
	go rl.evictionLoop()
	return rl
}

func (m *RateLimiter) Stop() {
	close(m.done)
}

func (m *RateLimiter) Limit(rules ...RateLimitRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if m == nil || len(rules) == 0 {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyValues, ok := m.bodyValues(w, r, rules)
			if !ok {
				return
			}

			now := time.Now().UTC()
			for _, rule := range rules {
				if rule.Limit <= 0 || rule.Window <= 0 {
					continue
				}

				rawKey := m.rawKey(r, rule, bodyValues)
				if rawKey == "" {
					continue
				}

				var allowed bool
				var count int64
				var retryAfter time.Duration
				if m.redis != nil {
					var err error
					allowed, count, retryAfter, err = m.allow(r.Context(), rawKey, rule, now)
					if err != nil {
						m.logger.Warn("rate limiter redis error, falling back to in-memory",
							zap.Error(err),
							zap.String("rule", rule.Name),
							zap.String("path", r.URL.Path),
						)
						allowed, count, retryAfter = m.allowInMemory(rawKey, rule, now)
					}
				} else {
					allowed, count, retryAfter = m.allowInMemory(rawKey, rule, now)
				}

				w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rule.Limit, 10))
				w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(max(rule.Limit-count, 0), 10))
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(now.Add(retryAfter).Unix(), 10))

				if !allowed {
					w.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
					response.Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests. Please try again later.", map[string]any{
						"retry_after_seconds": int64(retryAfter.Seconds()),
					})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *RateLimiter) allowInMemory(rawKey string, rule RateLimitRule, now time.Time) (bool, int64, time.Duration) {
	windowStart := now.Truncate(rule.Window)
	windowEnd := windowStart.Add(rule.Window)
	ttl := time.Until(windowEnd)
	if ttl <= 0 {
		ttl = rule.Window
	}
	keyHash := sha256.Sum256([]byte(rawKey))
	key := rule.Name + ":" + hex.EncodeToString(keyHash[:])
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.counters[key]; ok {
		if now.After(c.expiresAt) {
			c.count = 1
			c.expiresAt = windowEnd
		} else {
			c.count++
		}
		count := c.count
		return count <= rule.Limit, count, ttl
	}
	m.counters[key] = &inMemoryCounter{count: 1, expiresAt: windowEnd}
	return true, 1, ttl
}

func (m *RateLimiter) evictionLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.evictExpired()
		case <-m.done:
			return
		}
	}
}

func (m *RateLimiter) evictExpired() {
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, c := range m.counters {
		if now.After(c.expiresAt) {
			delete(m.counters, key)
		}
	}
}

func (m *RateLimiter) bodyValues(w http.ResponseWriter, r *http.Request, rules []RateLimitRule) (map[string]string, bool) {
	if !needsBody(rules) {
		return nil, true
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRateLimitBodySize+1))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body", nil)
		return nil, false
	}
	if len(body) > maxRateLimitBodySize {
		response.Error(w, http.StatusRequestEntityTooLarge, "REQUEST_BODY_TOO_LARGE", "Request body is too large", nil)
		return nil, false
	}

	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))

	if len(bytes.TrimSpace(body)) == 0 {
		return map[string]string{}, true
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return map[string]string{}, true
	}

	values := make(map[string]string, len(raw))
	for key, value := range raw {
		values[key] = strings.TrimSpace(fmt.Sprint(value))
	}

	return values, true
}

// incrementWithTTL increments the window counter and guarantees it carries an
// expiry, atomically.
//
// The previous INCR-then-EXPIRE pair was two round trips: a connection drop or
// failover between them left the key at TTL -1, incrementing forever. Once it
// passed the limit that caller received 429 permanently, with a Retry-After
// that never arrived and no recovery short of a manual DEL.
//
// The TTL is re-applied whenever it is missing rather than only on the first
// increment, so a key already stranded by the old code heals on its next hit.
var incrementWithTTL = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if redis.call("PTTL", KEYS[1]) < 0 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
`)

func (m *RateLimiter) allow(ctx context.Context, rawKey string, rule RateLimitRule, now time.Time) (bool, int64, time.Duration, error) {
	windowStart := now.Truncate(rule.Window)
	windowEnd := windowStart.Add(rule.Window)
	ttl := time.Until(windowEnd)
	if ttl <= 0 {
		ttl = rule.Window
	}

	keyHash := sha256.Sum256([]byte(rawKey))
	key := m.prefix + rule.Name + ":" + hex.EncodeToString(keyHash[:])

	count, err := incrementWithTTL.Run(
		ctx, m.redis, []string{key}, ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return false, 0, 0, err
	}

	if count > rule.Limit {
		redisTTL, err := m.redis.TTL(ctx, key).Result()
		if err != nil || redisTTL <= 0 {
			redisTTL = ttl
		}
		return false, count, redisTTL, nil
	}

	return true, count, ttl, nil
}

func (m *RateLimiter) rawKey(r *http.Request, rule RateLimitRule, bodyValues map[string]string) string {
	parts := []string{rule.Name}

	for _, field := range rule.KeyFields {
		switch field {
		case "ip":
			parts = append(parts, clientIPForRateLimit(r))
		case "path":
			parts = append(parts, r.URL.Path)
		case "user":
			parts = append(parts, "user="+rateLimitUserKey(r))
		}
	}

	for _, field := range rule.BodyFields {
		value := strings.ToLower(strings.TrimSpace(bodyValues[field]))
		if value == "" {
			value = "missing:" + field
		}
		parts = append(parts, field+"="+value)
	}

	return strings.Join(parts, "|")
}

// KnownKeyFields are the KeyFields rawKey understands. A field outside this set
// contributes nothing to the key, silently widening the bucket, so rules are
// validated at wiring time instead — see ValidateRules.
var KnownKeyFields = map[string]struct{}{
	"ip":   {},
	"path": {},
	"user": {},
}

// ValidateRules reports the first rule that would not key the way it declares.
// Called at route-registration time so a typo or a removed key field fails the
// build-out loudly rather than quietly granting a wider allowance.
func ValidateRules(rules ...RateLimitRule) error {
	for _, rule := range rules {
		for _, field := range rule.KeyFields {
			if _, ok := KnownKeyFields[field]; !ok {
				return fmt.Errorf(
					"rate limit rule %q declares unknown key field %q", rule.Name, field,
				)
			}
		}
		if rule.Limit <= 0 || rule.Window <= 0 {
			return fmt.Errorf("rate limit rule %q needs a positive Limit and Window", rule.Name)
		}
	}
	return nil
}

// rateLimitUserKey identifies the authenticated caller. It falls back to the
// client IP so a rule that reaches an unauthenticated request still buckets per
// caller rather than collapsing every anonymous caller into one counter.
func rateLimitUserKey(r *http.Request) string {
	authCtx, ok := r.Context().Value(appContracts.AuthContextKey).(appContracts.AuthContext)
	if ok && authCtx.UserID != "" {
		return authCtx.UserID
	}
	return "ip:" + clientIPForRateLimit(r)
}

func needsBody(rules []RateLimitRule) bool {
	for _, rule := range rules {
		if len(rule.BodyFields) > 0 {
			return true
		}
	}
	return false
}

func clientIPForRateLimit(r *http.Request) string {
	if meta, ok := requestmeta.FromContext(r.Context()); ok && meta.IPAddress != "" {
		return meta.IPAddress
	}
	// Fallback if metadata middleware didn't run
	return r.RemoteAddr
}
