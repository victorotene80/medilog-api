package bulksms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	httpclient "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	appconfig "github.com/victorotene80/medilog-api/internal/shared/config"
)

type DefaultMessagingService struct {
	cfg  appconfig.BulkSmsConfig
	http httpclient.HTTPService
}

func NewDefaultMessagingService(
	cfg appconfig.BulkSmsConfig,
	httpService httpclient.HTTPService,
) *DefaultMessagingService {
	return &DefaultMessagingService{
		cfg:  cfg,
		http: httpService,
	}
}

func (s *DefaultMessagingService) SendSMS(
	ctx context.Context,
	req BulkSmsRequest,
) (*SendMessageResponse, error) {
	to := strings.TrimSpace(req.To)
	from := strings.TrimSpace(req.From)
	body := strings.TrimSpace(req.Body)

	if to == "" {
		return nil, fmt.Errorf("recipient phone number is required")
	}

	if from == "" {
		return nil, fmt.Errorf("sender ID is required")
	}

	if body == "" {
		return nil, fmt.Errorf("message body is required")
	}

	if strings.TrimSpace(s.cfg.APIToken) == "" {
		return nil, fmt.Errorf("bulk SMS API token is required")
	}

	baseURL := strings.TrimRight(s.cfg.ProdBaseURL, "/")
	endpoint := "/" + strings.TrimLeft(s.cfg.SendMessagePath, "/")
	url := baseURL + endpoint

	payload := map[string]string{
		"from": from,
		"to":   to,
		"body": body,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	res, err := s.http.Do(ctx, httpclient.HTTPRequest{
		Method: http.MethodPost,
		URL:    url,
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", s.cfg.LegacyToken),
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		},
		Body: bytes.NewReader(payloadBytes),
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"bulk SMS request failed with status %d: %s",
			res.StatusCode,
			string(res.Body),
		)
	}

	var out SendMessageResponse
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
