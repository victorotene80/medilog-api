package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

type healthStatus struct {
	Statuses map[string]string `json:"statuses"`
}

func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redisClient,
	}
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	statuses := map[string]string{}
	healthy := true

	// Scan, not Raw alone: Raw only assembles the statement, so checking .Error
	// on it never contacts the database and this probe could not fail.
	var probe int
	if err := h.db.WithContext(ctx).Raw("SELECT 1").Scan(&probe).Error; err != nil {
		zap.L().Error("health check: postgres unhealthy", zap.Error(err))
		statuses["postgres"] = "unhealthy"
		healthy = false
	} else {
		statuses["postgres"] = "healthy"
	}

	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			zap.L().Error("health check: redis unhealthy", zap.Error(err))
			statuses["redis"] = "unhealthy"
			healthy = false
		} else {
			statuses["redis"] = "healthy"
		}
	} else {
		statuses["redis"] = "not configured"
	}

	code := http.StatusOK
	if !healthy {
		code = http.StatusServiceUnavailable
	}

	response.Success(w, code, "HEALTH_CHECK", "readiness check", &healthStatus{
		Statuses: statuses,
	})
}
