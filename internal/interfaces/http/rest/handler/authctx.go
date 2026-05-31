package handler

import (
	"context"
	"strconv"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
)

func UserIDFrom(ctx context.Context) (int64, bool) {
	authCtx, ok := ctx.Value(appContracts.AuthContextKey).(appContracts.AuthContext)
	if !ok || authCtx.UserID == "" {
		return 0, false
	}

	userID, err := strconv.ParseInt(authCtx.UserID, 10, 64)
	if err != nil || userID <= 0 {
		return 0, false
	}

	return userID, true
}

func AuthContextFrom(ctx context.Context) (appContracts.AuthContext, bool) {
	authCtx, ok := ctx.Value(appContracts.AuthContextKey).(appContracts.AuthContext)
	if !ok || authCtx.UserID == "" {
		return appContracts.AuthContext{}, false
	}

	return authCtx, true
}

func SessionIDFrom(ctx context.Context) (string, bool) {
	authCtx, ok := AuthContextFrom(ctx)
	if !ok {
		return "", false
	}

	return authCtx.SessionID, true
}
