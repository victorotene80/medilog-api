package twilio

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	httpclient "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	appconfig "github.com/victorotene80/medilog-api/internal/shared/config"
)

type DefaultTwilioService struct {
	cfg  appconfig.TwilioConfig
	http httpclient.HTTPService
}

func NewDefaultTwilioService(
	cfg appconfig.TwilioConfig,
	httpService httpclient.HTTPService,
) *DefaultTwilioService {
	return &DefaultTwilioService{
		cfg:  cfg,
		http: httpService,
	}
}

func (s *DefaultTwilioService) SendSMS(
	ctx context.Context,
	req SendMessageRequest,
) (*SendMessageResult, error) {
	return s.sendMessage(ctx, sendMessageOptions{
		To:          req.To,
		Body:        req.Body,
		From:        s.cfg.FromNumber,
		UseWhatsApp: false,
	})
}

func (s *DefaultTwilioService) SendWhatsApp(
	ctx context.Context,
	req SendMessageRequest,
) (*SendMessageResult, error) {
	return s.sendMessage(ctx, sendMessageOptions{
		To:          req.To,
		Body:        req.Body,
		From:        s.cfg.WhatsAppFromNumber,
		UseWhatsApp: true,
	})
}

type sendMessageOptions struct {
	To          string
	Body        string
	From        string
	UseWhatsApp bool
}

func (s *DefaultTwilioService) sendMessage(
	ctx context.Context,
	opt sendMessageOptions,
) (*SendMessageResult, error) {
	to := strings.TrimSpace(opt.To)
	body := opt.Body

	if to == "" {
		return nil, fmt.Errorf("recipient phone number is required")
	}

	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("message body is required")
	}

	form := s.buildMessageForm(to, body, opt.From, opt.UseWhatsApp)

	return s.send(ctx, form)
}

func (s *DefaultTwilioService) buildMessageForm(
	to string,
	body string,
	from string,
	useWhatsApp bool,
) url.Values {
	form := url.Values{}
	form.Set("To", formatRecipient(to, useWhatsApp))
	form.Set("Body", body)

	if strings.TrimSpace(s.cfg.MessagingServiceSID) != "" {
		form.Set("MessagingServiceSid", strings.TrimSpace(s.cfg.MessagingServiceSID))
		return form
	}

	form.Set("From", formatRecipient(from, useWhatsApp))

	return form
}

func (s *DefaultTwilioService) send(
	ctx context.Context,
	form url.Values,
) (*SendMessageResult, error) {
	res, err := s.http.Do(ctx, httpclient.HTTPRequest{
		Method: http.MethodPost,
		URL:    s.messagesURL(),
		Headers: map[string]string{
			"Authorization": basicAuth(s.cfg.AccountSID, s.cfg.AuthToken),
			"Content-Type":  "application/x-www-form-urlencoded",
			"Accept":        "application/json",
		},
		Body: form.Encode(),
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, parseTwilioError(res)
	}

	var out SendMessageResult
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *DefaultTwilioService) messagesURL() string {
	return strings.ReplaceAll(
		s.cfg.MessagesURL,
		"{accountSid}",
		url.PathEscape(s.cfg.AccountSID),
	)
}

func formatRecipient(value string, useWhatsApp bool) string {
	value = strings.TrimSpace(value)

	if !useWhatsApp {
		return value
	}

	if strings.HasPrefix(value, "whatsapp:") {
		return value
	}

	return "whatsapp:" + value
}

func basicAuth(username, password string) string {
	raw := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

func parseTwilioError(res *httpclient.HTTPResult) error {
	if res == nil {
		return fmt.Errorf("twilio request failed")
	}

	var twilioErr twilioErrorResponse
	if len(res.Body) > 0 {
		if err := json.Unmarshal(res.Body, &twilioErr); err == nil && twilioErr.Message != "" {
			return fmt.Errorf(
				"twilio request failed: status=%d code=%d message=%s",
				res.StatusCode,
				twilioErr.Code,
				twilioErr.Message,
			)
		}
	}

	return fmt.Errorf(
		"twilio request failed: status=%d body=%s",
		res.StatusCode,
		string(res.Body),
	)
}
