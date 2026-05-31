package entities

import "time"

type UserAuthProvider struct {
	ID          int64
	UserID      int64
	Provider    string
	ProviderUID string
	Email       *string
	IsPrimary   bool
	LinkedAt    time.Time
}

func NewUserAuthProvider(
	userID int64,
	provider string,
	providerUID string,
	email *string,
	isPrimary bool,
) *UserAuthProvider {
	return &UserAuthProvider{
		UserID:      userID,
		Provider:    provider,
		ProviderUID: providerUID,
		Email:       email,
		IsPrimary:   isPrimary,
		LinkedAt:    time.Now(),
	}
}
