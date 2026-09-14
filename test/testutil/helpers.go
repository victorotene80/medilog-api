package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
)

func SetupTestContext(userID, sessionID string) context.Context {
	ctx := context.Background()
	authCtx := appContracts.AuthContext{
		UserID:    userID,
		SessionID: sessionID,
	}
	return context.WithValue(ctx, appContracts.AuthContextKey, authCtx)
}

func RandomEmail() string {
	return fmt.Sprintf("test-%s@example.com", RandomString(8))
}

func RandomPhone() string {
	return fmt.Sprintf("+1555%s", RandomString(7))
}

func RandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)[:n]
}

func MustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

func PtrString(s string) *string {
	return &s
}

func PtrInt64(i int64) *int64 {
	return &i
}

func PtrInt(i int) *int {
	return &i
}

func PtrTime(t time.Time) *time.Time {
	return &t
}

func FixedClock() func() time.Time {
	fixed := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return fixed }
}

func FutureTime(d time.Duration) time.Time {
	return time.Now().UTC().Add(d)
}

func PastTime(d time.Duration) time.Time {
	return time.Now().UTC().Add(-d)
}
