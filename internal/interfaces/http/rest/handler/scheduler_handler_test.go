package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubReminderRunner struct {
	calls int
	err   error
}

func (s *stubReminderRunner) Tick(context.Context) error {
	s.calls++
	return s.err
}

func TestSchedulerHandler_TickReminders_Authorization(t *testing.T) {
	const token = "s3cret-tick-token"

	tests := []struct {
		name          string
		configured    string
		header        string
		sendHeader    bool
		wantStatus    int
		wantRunnerHit bool
	}{
		{
			name:          "correct token runs the scan",
			configured:    token,
			header:        token,
			sendHeader:    true,
			wantStatus:    http.StatusOK,
			wantRunnerHit: true,
		},
		{
			name:       "wrong token is rejected",
			configured: token,
			header:     "wrong-token-same-len",
			sendHeader: true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing header is rejected",
			configured: token,
			sendHeader: false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			// A prefix of the real token must not pass. The length guard in
			// authorized() exists for this, since ConstantTimeCompare would
			// otherwise never see mismatched lengths.
			name:       "token prefix is rejected",
			configured: token,
			header:     token[:5],
			sendHeader: true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			// Defence in depth: the route is not registered at all when the
			// token is unset, but the handler must not fall open if it were.
			name:       "unset token denies everything",
			configured: "",
			header:     "",
			sendHeader: true,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &stubReminderRunner{}
			h := NewSchedulerHandler(runner, tt.configured, nil)

			req := httptest.NewRequest(http.MethodPost, "/internal/scheduler/reminders/tick", nil)
			if tt.sendHeader {
				req.Header.Set("X-Scheduler-Token", tt.header)
			}
			rec := httptest.NewRecorder()

			h.TickReminders(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantRunnerHit {
				assert.Equal(t, 1, runner.calls, "scan should have run")
			} else {
				assert.Zero(t, runner.calls, "scan must not run for an unauthorized caller")
			}
		})
	}
}

func TestSchedulerHandler_TickReminders_RunnerError(t *testing.T) {
	const token = "s3cret-tick-token"

	runner := &stubReminderRunner{err: errors.New("scan blew up")}
	h := NewSchedulerHandler(runner, token, nil)

	req := httptest.NewRequest(http.MethodPost, "/internal/scheduler/reminders/tick", nil)
	req.Header.Set("X-Scheduler-Token", token)
	rec := httptest.NewRecorder()

	h.TickReminders(rec, req)

	// Cloud Scheduler retries on 5xx, which is what we want: a failed scan
	// should be attempted again rather than silently skipped.
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, 1, runner.calls)
}
