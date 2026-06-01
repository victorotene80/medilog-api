package google

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/shared/config"
	"google.golang.org/api/idtoken"
)

type AuthService struct {
	clientID string
}

func NewAuthService(cfg config.GoogleConfig) *AuthService {
	return &AuthService{
		clientID: cfg.ClientID,
	}
}

func (s *AuthService) VerifyIDToken(
	ctx context.Context,
	token string,
) (*contracts.GoogleUserInfo, error) {
	payload, err := idtoken.Validate(ctx, token, s.clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid google id token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)
	firstName, _ := payload.Claims["given_name"].(string)
	lastName, _ := payload.Claims["family_name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	if !emailVerified {
		return nil, fmt.Errorf("google email is not verified")
	}

	return &contracts.GoogleUserInfo{
		GoogleID:      payload.Subject,
		Email:         email,
		EmailVerified: true,
		Name:          name,
		FirstName:     firstName,
		LastName:      lastName,
		PictureURL:    picture,
	}, nil
}
