package bootstrap

import (
	"strings"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
	aiinfra "github.com/victorotene80/medilog-api/internal/infrastructure/services/ai"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/bulksms"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/claude"
	googleauth "github.com/victorotene80/medilog-api/internal/infrastructure/services/google"
	httpclient "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/sms"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/telnyx"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/twilio"
	"github.com/victorotene80/medilog-api/internal/shared/config"
)

type ExternalServices struct {
	GoogleAuth contracts.GoogleAuthService
	SMSSender  contracts.SMSSender
	AIModel    contracts.AIModelService
}

func initializeExternalServices(
	cfg *config.Config,
	httpService httpclient.HTTPService,
) ExternalServices {
	return ExternalServices{
		GoogleAuth: googleauth.NewAuthService(cfg.Google),
		SMSSender:  buildSMSSender(cfg, httpService),
		AIModel:    buildAIModelService(cfg),
	}
}

func buildAIModelService(cfg *config.Config) contracts.AIModelService {
	if !cfg.AI.Enabled {
		return aiinfra.NewNoopModelService("AI is disabled")
	}

	switch strings.ToLower(strings.TrimSpace(cfg.AI.Provider)) {
	case "", "claude", "anthropic":
		if strings.TrimSpace(cfg.Claude.APIKey) == "" {
			return aiinfra.NewNoopModelService("Claude API key is missing")
		}

		// A full chat reply routinely outlasts the 10s shared client used for
		// SMS, so Claude gets its own client with a longer budget.
		claudeHTTP := initializeHTTPClient(config.HTTPConfig{Timeout: cfg.Claude.Timeout})
		claudeService := claude.NewDefaultClaudeService(cfg.Claude, claudeHTTP)
		return claude.NewClaudeModelAdapter(claudeService)
	default:
		return aiinfra.NewNoopModelService("unsupported provider " + cfg.AI.Provider)
	}
}

func buildSMSSender(cfg *config.Config, httpService httpclient.HTTPService) contracts.SMSSender {
	if !cfg.SMS.Enabled {
		return sms.NewUnavailableSender("SMS is disabled (SMS_ENABLED=false)")
	}

	var providers []contracts.SMSSender

	if strings.TrimSpace(cfg.Telnyx.APIKey) != "" {
		telnyxService := telnyx.NewDefaultTelnyxService(cfg.Telnyx, httpService)
		providers = append(providers, telnyx.NewSenderAdapter(telnyxService))
	}

	if strings.TrimSpace(cfg.Twilio.AccountSID) != "" &&
		strings.TrimSpace(cfg.Twilio.AuthToken) != "" {
		twilioService := twilio.NewDefaultTwilioService(cfg.Twilio, httpService)
		providers = append(providers, twilio.NewSenderAdapter(twilioService))
	}

	if strings.TrimSpace(cfg.BulkSms.APIToken) != "" {
		bulkSMSService := bulksms.NewDefaultMessagingService(cfg.BulkSms, httpService)
		providers = append(providers, bulksms.NewSenderAdapter(bulkSMSService, *cfg))
	}

	if len(providers) == 0 {
		// SMS was switched on but no provider has credentials. Failing here is the
		// point: this is the misconfiguration most likely to reach production, and
		// it previously presented as every OTP succeeding.
		return sms.NewUnavailableSender(
			"SMS_ENABLED=true but no provider is configured (set Telnyx, Twilio or BulkSMS credentials)",
		)
	}

	return sms.NewChainSender(providers...)
}
