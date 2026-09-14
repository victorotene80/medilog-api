package middleware

import (
	"context"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
)

type mockAuthService struct {
	authenticateResult     appContracts.AuthContext
	authenticateErr        error
	checkAdminAccessResult error
	checkUserAccessResult  error
	checkUserActiveResult  error
}

func (m *mockAuthService) Authenticate(_ context.Context, _ string) (appContracts.AuthContext, error) {
	return m.authenticateResult, m.authenticateErr
}

func (m *mockAuthService) CheckUserAccess(_ context.Context, _ int64) error {
	return m.checkUserAccessResult
}

func (m *mockAuthService) CheckUserActive(_ context.Context, _ int64) error {
	return m.checkUserActiveResult
}

func (m *mockAuthService) CheckAdminAccess(_ context.Context, _ int64) error {
	return m.checkAdminAccessResult
}
