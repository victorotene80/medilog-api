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
	"time"

	"github.com/go-redis/redis/v8"
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

type RateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
	prefix string
}

func NewRateLimiter(redisClient *redis.Client, logger *zap.Logger) *RateLimiter {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RateLimiter{
		redis:  redisClient,
		logger: logger,
		prefix: "rate_limit:",
	}
}

func (m *RateLimiter) Limit(rules ...RateLimitRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if m == nil || m.redis == nil || len(rules) == 0 {
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

				allowed, count, retryAfter, err := m.allow(r.Context(), rawKey, rule, now)
				if err != nil {
					m.logger.Error("rate limiter redis error",
						zap.Error(err),
						zap.String("rule", rule.Name),
						zap.String("path", r.URL.Path),
					)
					response.Error(w, http.StatusServiceUnavailable, "RATE_LIMIT_UNAVAILABLE", "Rate limiting is temporarily unavailable", nil)
					return
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

func (m *RateLimiter) allow(ctx context.Context, rawKey string, rule RateLimitRule, now time.Time) (bool, int64, time.Duration, error) {
	windowStart := now.Truncate(rule.Window)
	windowEnd := windowStart.Add(rule.Window)
	ttl := time.Until(windowEnd)
	if ttl <= 0 {
		ttl = rule.Window
	}

	keyHash := sha256.Sum256([]byte(rawKey))
	key := m.prefix + rule.Name + ":" + hex.EncodeToString(keyHash[:])

	count, err := m.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, 0, err
	}

	if count == 1 {
		if err := m.redis.Expire(ctx, key, ttl).Err(); err != nil {
			return false, 0, 0, err
		}
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
		case "device":
			parts = append(parts, deviceFingerprintForRateLimit(r))
		case "path":
			parts = append(parts, r.URL.Path)
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
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	return r.RemoteAddr
}

func deviceFingerprintForRateLimit(r *http.Request) string {
	if meta, ok := requestmeta.FromContext(r.Context()); ok && meta.DeviceFingerprint != "" {
		return meta.DeviceFingerprint
	}
	if fp := strings.TrimSpace(r.Header.Get("X-Device-Fingerprint")); fp != "" {
		return fp
	}
	return "unknown-device"
}
