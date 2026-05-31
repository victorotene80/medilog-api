package telnyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	httpclient "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	appconfig "github.com/victorotene80/medilog-api/internal/shared/config"
)

type DefaultTelnyxService struct {
	cfg  appconfig.TelnyxConfig
	http httpclient.HTTPService
}

func NewDefaultTelnyxService(
	cfg appconfig.TelnyxConfig,
	httpService httpclient.HTTPService,
) *DefaultTelnyxService {
	return &DefaultTelnyxService{
		cfg:  cfg,
		http: httpService,
	}
}

func (s *DefaultTelnyxService) SendSMS(
	ctx context.Context,
	req SendMessageRequest,
) (*SendMessageResult, error) {
	to := strings.TrimSpace(req.To)
	body := strings.TrimSpace(req.Body)

	if to == "" {
		return nil, fmt.Errorf("recipient phone number is required")
	}

	if body == "" {
		return nil, fmt.Errorf("message body is required")
	}

	payload, err := s.buildPayload(to, body)
	if err != nil {
		return nil, err
	}

	return s.send(ctx, payload)
}

func (s *DefaultTelnyxService) buildPayload(to, body string) (map[string]any, error) {
	from := strings.TrimSpace(s.cfg.FromNumber)
	if from == "" {
		return nil, fmt.Errorf("telnyx from number is required")
	}

	return map[string]any{
		"from": from,
		"to":   to,
		"text": body,
	}, nil
}

func (s *DefaultTelnyxService) send(
	ctx context.Context,
	payload map[string]any,
) (*SendMessageResult, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("telnyx: failed to marshal request: %w", err)
	}

	res, err := s.http.Do(ctx, httpclient.HTTPRequest{
		Method: http.MethodPost,
		URL:    s.cfg.MessagesURL,
		Headers: map[string]string{
			"Authorization": "Bearer " + s.cfg.APIKey,
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		},
		Body: string(body),
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, parseTelnyxError(res)
	}

	var wrapper telnyxDataWrapper
	if err := json.Unmarshal(res.Body, &wrapper); err != nil {
		return nil, fmt.Errorf("telnyx: failed to decode response: %w", err)
	}

	return &wrapper.Data, nil
}

func parseTelnyxError(res *httpclient.HTTPResult) error {
	if res == nil {
		return fmt.Errorf("telnyx: request failed")
	}

	if len(res.Body) > 0 {
		var errResp telnyxErrorResponse
		if err := json.Unmarshal(res.Body, &errResp); err == nil && len(errResp.Errors) > 0 {
			e := errResp.Errors[0]
			return fmt.Errorf(
				"telnyx: request failed: status=%d code=%s title=%s detail=%s",
				res.StatusCode,
				e.Code,
				e.Title,
				e.Detail,
			)
		}
	}

	return fmt.Errorf(
		"telnyx: request failed: status=%d body=%s",
		res.StatusCode,
		string(res.Body),
	)
}
