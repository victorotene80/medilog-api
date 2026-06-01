package bootstrap

import (
	"strings"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
	aiinfra "github.com/victorotene80/medilog-api/internal/infrastructure/services/ai"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/claude"
	googleauth "github.com/victorotene80/medilog-api/internal/infrastructure/services/google"
	httpclient "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/sms"
	"github.com/victorotene80/medilog-api/internal/infrastructure/services/telnyx"
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
		AIModel:    buildAIModelService(cfg, httpService),
	}
}

func buildAIModelService(cfg *config.Config, httpService httpclient.HTTPService) contracts.AIModelService {
	if !cfg.AI.Enabled {
		return aiinfra.NewNoopModelService("AI is disabled")
	}

	switch strings.ToLower(strings.TrimSpace(cfg.AI.Provider)) {
	case "", "claude", "anthropic":
		if strings.TrimSpace(cfg.Claude.APIKey) == "" {
			return aiinfra.NewNoopModelService("Claude API key is missing")
		}

		claudeService := claude.NewDefaultClaudeService(cfg.Claude, httpService)
		return claude.NewClaudeModelAdapter(claudeService)
	default:
		return aiinfra.NewNoopModelService("unsupported provider " + cfg.AI.Provider)
	}
}

func buildSMSSender(cfg *config.Config, httpService httpclient.HTTPService) contracts.SMSSender {
	if !cfg.SMS.Enabled {
		return sms.NewNoopSender()
	}

	var providers []contracts.SMSSender

	if strings.TrimSpace(cfg.Telnyx.APIKey) != "" {
		telnyxService := telnyx.NewDefaultTelnyxService(cfg.Telnyx, httpService)
		providers = append(providers, telnyx.NewSenderAdapter(telnyxService))
	}

	//if strings.TrimSpace(cfg.Twilio.AccountSID) != "" &&
	//	strings.TrimSpace(cfg.Twilio.AuthToken) != "" {
	//	twilioService := twilio.NewDefaultTwilioService(cfg.Twilio, httpService)
	//	providers = append(providers, twilio.NewSenderAdapter(twilioService))
	//}

	//if strings.TrimSpace(cfg.BulkSms.APIToken) != "" {
	//	bulkSMSService := bulksms.NewDefaultMessagingService(cfg.BulkSms, httpService)
	//	providers = append(providers, bulksms.NewSenderAdapter(bulkSMSService, *cfg))
	//}

	if len(providers) == 0 {
		return sms.NewNoopSender()
	}

	return sms.NewChainSender(providers...)
}

/*
func buildSMSSender(cfg *config.Config, httpService httpclient.HTTPService) contracts.SMSSender {
	if !cfg.SMS.Enabled {
		return sms.NewNoopSender()
	}

	bulkSMSService := bulksms.NewDefaultMessagingService(cfg.BulkSms, httpService)
	bulkSMSSender := bulksms.NewSenderAdapter(bulkSMSService, *cfg)

	twilioService := twilio.NewDefaultTwilioService(cfg.Twilio, httpService)
	twilioSender := twilio.NewSenderAdapter(twilioService)

	maxFailures := cfg.SMS.MaxBulkFailures
	if maxFailures <= 0 {
		maxFailures = 3
	}

	return services.NewRoutedSMSSender(bulkSMSSender, twilioSender, maxFailures)
}
*/
