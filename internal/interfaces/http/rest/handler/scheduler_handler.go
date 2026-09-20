package handler

import (
	"context"
	"crypto/subtle"
	"net/http"

	"github.com/victorotene80/medilog-api/internal/application"
	"go.uber.org/zap"
)

// ReminderRunner is satisfied by *scheduler.ReminderScheduler. It is declared
// here rather than imported so the interfaces layer keeps no dependency on
// infrastructure; bootstrap supplies the concrete type.
type ReminderRunner interface {
	Tick(ctx context.Context) error
}

// SchedulerHandler exposes the reminder scan that used to run on an in-process
// ticker. Cloud Run throttles a container's CPU to near zero between requests
// and stops it entirely at zero instances, so periodic work there has to be
// driven from outside — Cloud Scheduler calls this once a minute.
type SchedulerHandler struct {
	runner ReminderRunner
	token  string
	logger *zap.Logger
}

func NewSchedulerHandler(runner ReminderRunner, token string, logger *zap.Logger) *SchedulerHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SchedulerHandler{runner: runner, token: token, logger: logger}
}

// TickReminders runs one reminder scan. POST, 200 on success, 401 without a
// valid X-Scheduler-Token.
//
// Deliberately carries no swagger annotations. docs/swagger.json is the
// contract the mobile team codes against and is served publicly at /swagger —
// listing a machine-to-machine endpoint there advertises it to everyone who
// opens the UI while giving its only real caller, Cloud Scheduler, nothing it
// does not already have in cloudbuild.yaml.
func (h *SchedulerHandler) TickReminders(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		// Deliberately no detail about which check failed, and the same
		// response whether the token is wrong or unset.
		logAndRespond(w, http.StatusUnauthorized, application.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	// r.Context() rather than a fresh one: if Cloud Scheduler hangs up or Cloud
	// Run cuts the request short, the scan should stop with it. Abandoning a
	// half-finished scan is safe — the next tick's catch-up window covers the
	// same ground, and the dedupe index stops the overlap creating duplicates.
	if err := h.runner.Tick(r.Context()); err != nil {
		logAndRespond(w, http.StatusInternalServerError, application.CodeSchedulerTickFailed, "Reminder scan failed", err)
		return
	}

	writeEmptySuccess(w, http.StatusOK, application.CodeSchedulerTickCompleted, "Reminder scan completed")
}

// authorized compares in constant time so the token cannot be recovered a byte
// at a time from response timing. An unset token denies everything: the scan
// writes notifications, so an open endpoint would let anyone flood an inbox.
func (h *SchedulerHandler) authorized(r *http.Request) bool {
	if h.token == "" {
		return false
	}

	supplied := r.Header.Get("X-Scheduler-Token")

	// ConstantTimeCompare returns 0 on a length mismatch without comparing, so
	// the length check is what keeps *that* branch constant time too.
	if len(supplied) != len(h.token) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(supplied), []byte(h.token)) == 1
}
