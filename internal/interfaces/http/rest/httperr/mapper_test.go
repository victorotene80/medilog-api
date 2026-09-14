package httperr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

// A write that matched no rows is a 404, not a 500. Repositories used to signal
// this with gorm.ErrRecordNotFound, which nothing above infrastructure could
// test for, so a concurrent update or delete fell through to the default branch.
func TestStatusFrom_RepositoryNotFound_Is404(t *testing.T) {
	assert.Equal(t, http.StatusNotFound, StatusFrom(repository.ErrNotFound))
}

// The signal has to survive the fmt.Errorf("...: %w", err) wrapping that every
// application handler applies on the way out.
func TestStatusFrom_WrappedRepositoryNotFound_Is404(t *testing.T) {
	wrapped := fmt.Errorf("update visit: %w", repository.ErrNotFound)
	assert.Equal(t, http.StatusNotFound, StatusFrom(wrapped))
}

// Genuine infrastructure failures must still be 500 — the point of the default
// branch is that unknown errors are not misreported as client errors.
func TestStatusFrom_UnknownError_Is500(t *testing.T) {
	assert.Equal(t, http.StatusInternalServerError,
		StatusFrom(errors.New("dial tcp: connection refused")))
	assert.Equal(t, http.StatusInternalServerError, StatusFrom(nil))
}

func TestStatusFrom_AppErrorsTakePriority(t *testing.T) {
	assert.Equal(t, http.StatusConflict, StatusFrom(application.NewConflict("exists")))
	assert.Equal(t, http.StatusNotFound, StatusFrom(application.NewNotFound("gone")))
	assert.Equal(t, http.StatusPaymentRequired, StatusFrom(application.NewQuotaExceeded("spent")))
}
