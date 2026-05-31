package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/google/response"
	httpclient "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	"github.com/victorotene80/medilog-api/internal/shared/config"
)

type AuthService struct {
	httpService  httpclient.HTTPService
	clientID     string
	tokenInfoURL string
	userInfoURL  string
	oAuthBaseURL string
}

func NewAuthService(
	httpService httpclient.HTTPService,
	cfg config.GoogleConfig,
) *AuthService {
	return &AuthService{
		httpService:  httpService,
		clientID:     cfg.ClientID,
		tokenInfoURL: cfg.TokenInfoURL,
		userInfoURL:  cfg.UserInfoURL,
		oAuthBaseURL: cfg.OAuthBaseURL,
	}
}

func (s *AuthService) VerifyIDToken(
	ctx context.Context,
	idToken string,
) (*contracts.GoogleUserInfo, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, fmt.Errorf("google id token is required")
	}

	result, err := s.httpService.Do(ctx, httpclient.HTTPRequest{
		Method: http.MethodGet,
		URL:    s.tokenInfoURL,
		QueryParams: map[string]string{
			"id_token": idToken,
		},
	})
	if err != nil {
		return nil, err
	}

	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"google token verification failed: status=%d body=%s",
			result.StatusCode,
			string(result.Body),
		)
	}

	var tokenInfo response.TokenInfoResponse
	if err := json.Unmarshal(result.Body, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode google token response: %w", err)
	}

	if tokenInfo.Audience != s.clientID {
		return nil, fmt.Errorf("invalid google token audience")
	}

	if tokenInfo.Issuer != "accounts.google.com" &&
		tokenInfo.Issuer != "https://accounts.google.com" {
		return nil, fmt.Errorf("invalid google token issuer")
	}

	if tokenInfo.Subject == "" {
		return nil, fmt.Errorf("missing google subject")
	}

	if tokenInfo.Email == "" {
		return nil, fmt.Errorf("missing google email")
	}

	if tokenInfo.EmailVerified != "true" {
		return nil, fmt.Errorf("google email is not verified")
	}

	return &contracts.GoogleUserInfo{
		GoogleID:      tokenInfo.Subject,
		Email:         tokenInfo.Email,
		EmailVerified: true,
		Name:          tokenInfo.Name,
		FirstName:     tokenInfo.GivenName,
		LastName:      tokenInfo.FamilyName,
		PictureURL:    tokenInfo.Picture,
	}, nil
}
