package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Load() (*Config, error) {
	security, err := loadSecurity()
	if err != nil {
		return nil, err
	}

	database, err := loadDatabase()
	if err != nil {
		return nil, err
	}

	redisCfg, err := loadRedis()
	if err != nil {
		return nil, err
	}

	httpCfg, err := loadHTTP()
	if err != nil {
		return nil, err
	}

	messaging := loadMessaging()

	googleCfg, err := loadGoogle()
	if err != nil {
		return nil, err
	}

	claude, err := loadClaude()
	if err != nil {
		return nil, err
	}

	ai, err := loadAI()
	if err != nil {
		return nil, err
	}

	otp, err := loadOTP()
	if err != nil {
		return nil, err
	}

	telemetry := loadTelemetry()

	twilio, err := loadTwilio()
	if err != nil {
		return nil, err
	}

	bulkSms, err := loadBulkSms()
	if err != nil {
		return nil, err
	}

	sms, err := loadSms()
	if err != nil {
		return nil, err
	}

	telnyx, err := loadTelnyx()
	if err != nil {
		return nil, err
	}

	app := loadApp()

	scheduler := loadScheduler()

	return &Config{
		App:       app,
		Security:  security,
		Database:  database,
		Redis:     redisCfg,
		Messaging: messaging,
		HTTP:      httpCfg,
		Google:    googleCfg,
		AI:        ai,
		Claude:    claude,
		OTP:       otp,
		Telemetry: telemetry,
		Twilio:    twilio,
		BulkSms:   bulkSms,
		SMS:       sms,
		Telnyx:    telnyx,
		Scheduler: scheduler,
	}, nil
}

func loadScheduler() SchedulerConfig {
	return SchedulerConfig{
		// Off by default: enabling it starts writing notifications, which is a
		// deliberate operational decision rather than something a fresh
		// environment should begin doing on its own.
		Enabled:       getBoolOrDefault("SCHEDULER_ENABLED", false),
		PollInterval:  getDurationOrDefault("SCHEDULER_POLL_INTERVAL", time.Minute),
		CatchupWindow: getDurationOrDefault("SCHEDULER_CATCHUP_WINDOW", 6*time.Hour),
		LeadTime:      getDurationOrDefault("SCHEDULER_LEAD_TIME", 5*time.Minute),
		BatchSize:     getIntOrDefault("SCHEDULER_BATCH_SIZE", 500),
	}
}

func loadAI() (AIConfig, error) {
	contextWindowTokens, err := getPositiveInt("AI_CONTEXT_WINDOW_TOKENS")
	if err != nil {
		return AIConfig{}, err
	}

	maxContextMessages, err := getPositiveInt("AI_MAX_CONTEXT_MESSAGES")
	if err != nil {
		return AIConfig{}, err
	}

	summaryTokenThreshold, err := getPositiveInt("AI_SUMMARY_TOKEN_THRESHOLD")
	if err != nil {
		return AIConfig{}, err
	}

	return AIConfig{
		Enabled:               getBoolOrDefault("AI_ENABLED", true),
		Provider:              getStringOrDefault("AI_PROVIDER", "claude"),
		ContextWindowTokens:   contextWindowTokens,
		MaxContextMessages:    maxContextMessages,
		SummaryTokenThreshold: summaryTokenThreshold,
	}, nil
}

func loadTelnyx() (TelnyxConfig, error) {
	apiKey, err := getString("TELNYX_API_KEY")
	if err != nil {
		return TelnyxConfig{}, err
	}

	return TelnyxConfig{
		APIKey: apiKey,

		FromNumber:         getStringOrDefault("TELNYX_FROM_NUMBER", ""),
		MessagingProfileID: getStringOrDefault("TELNYX_MESSAGING_PROFILE_ID", ""),

		MessagesURL: getStringOrDefault(
			"TELNYX_MESSAGES_URL",
			"https://api.telnyx.com/v2/messages",
		),
	}, nil
}

func loadTwilio() (TwilioConfig, error) {
	accountSID, err := getString("TWILIO_ACCOUNT_SID")
	if err != nil {
		return TwilioConfig{}, err
	}

	authToken, err := getString("TWILIO_AUTH_TOKEN")
	if err != nil {
		return TwilioConfig{}, err
	}

	baseURL := getStringOrDefault(
		"TWILIO_BASE_URL",
		"https://api.twilio.com",
	)

	return TwilioConfig{
		AccountSID: accountSID,
		AuthToken:  authToken,

		FromNumber:          getStringOrDefault("TWILIO_FROM_NUMBER", ""),
		WhatsAppFromNumber:  getStringOrDefault("TWILIO_WHATSAPP_FROM_NUMBER", ""),
		MessagingServiceSID: getStringOrDefault("TWILIO_MESSAGING_SERVICE_SID", ""),

		BaseURL: baseURL,

		MessagesURL: getStringOrDefault(
			"TWILIO_MESSAGES_URL",
			strings.TrimRight(baseURL, "/")+"/2010-04-01/Accounts/{accountSid}/Messages.json",
		),
	}, nil
}

func loadSms() (SMSConfig, error) {
	isActive := getBoolOrDefault("SMS_ENABLED", true)
	maxRetries, err := getInt("SMS_MAX_BULK_FAILURES")
	if err != nil {
		return SMSConfig{}, err
	}

	return SMSConfig{
		Enabled:         isActive,
		MaxBulkFailures: maxRetries,
	}, nil
}

func loadBulkSms() (BulkSmsConfig, error) {
	apiToken, err := getString("BULK_SMS_API_TOKEN")
	if err != nil {
		return BulkSmsConfig{}, err
	}

	sender, err := getString("BULK_SMS_SENDER")
	if err != nil {
		return BulkSmsConfig{}, err
	}

	return BulkSmsConfig{
		ProdBaseURL: getStringOrDefault(
			"BULK_SMS_PROD_BASEURL",
			"https://www.bulksmsnigeria.com/api/v2",
		),

		TestBaseURL: getStringOrDefault(
			"BULK_SMS_TEST_BASEURL",
			"https://www.bulksmsnigeria.com/api/sandbox/v2",
		),

		Sender: sender,

		SendMessagePath: getStringOrDefault(
			"BULK_SMS_SEND_MESSAGE",
			"/sms",
		),

		CheckBalancePath: getStringOrDefault(
			"BULK_SMS_CHECK_BALANCE",
			"",
		),

		DeliveryReportPath: getStringOrDefault(
			"BULK_SMS_GET_DELIVERY_REPORT",
			"",
		),

		APIToken: apiToken,

		LegacyToken: getStringOrDefault(
			"BULK_SMS_API_LEGACY_TOKEN",
			"",
		),
	}, nil
}

func loadGoogle() (GoogleConfig, error) {
	clientID, err := getString("GOOGLE_CLIENT_ID")
	if err != nil {
		return GoogleConfig{}, err
	}

	return GoogleConfig{
		ClientID: clientID,

		TokenInfoURL: getStringOrDefault(
			"GOOGLE_TOKEN_INFO_URL",
			"https://oauth2.googleapis.com/tokeninfo",
		),

		UserInfoURL: getStringOrDefault(
			"GOOGLE_USER_INFO_URL",
			"https://www.googleapis.com/oauth2/v3/userinfo",
		),

		OAuthBaseURL: getStringOrDefault(
			"GOOGLE_OAUTH_BASE_URL",
			"https://accounts.google.com/o/oauth2/v2/auth",
		),
	}, nil
}

func loadOTP() (OTPConfig, error) {
	return OTPConfig{
		Length:     getIntOrDefault("OTP_LENGTH", 6),
		TTLMins:    getIntOrDefault("OTP_TTL_MINS", 10),
		BcryptCost: getIntOrDefault("OTP_BCRYPT_COST", 0), // 0 → bcrypt.DefaultCost in service
	}, nil

}

func loadClaude() (ClaudeConfig, error) {
	apiKey := getStringOrDefault("CLAUDE_API_KEY", "")
	if apiKey == "" {
		apiKey = getStringOrDefault("ANTHROPIC_API_KEY", "")
	}

	return ClaudeConfig{
		APIKey: apiKey,
		BaseURL: getStringOrDefault(
			"CLAUDE_BASE_URL",
			"https://api.anthropic.com",
		),
		Model: getStringOrDefault(
			"CLAUDE_MODEL",
			getStringOrDefault("CLAUDE_BALANCED_MODEL", "claude-haiku-4-5-20251001"),
		),
		MaxTokens:   getIntOrDefault("CLAUDE_MAX_TOKENS", 1024),
		APIVersion:  getStringOrDefault("CLAUDE_API_VERSION", "2023-06-01"),
		WorkspaceID: getStringOrDefault("CLAUDE_WORKSPACE_ID", ""),
	}, nil
}

func loadSecurity() (SecurityConfig, error) {
	pepper, err := getString("SESSION_PEPPER")
	if err != nil {
		return SecurityConfig{}, err
	}

	jwtSecret, err := getString("JWT_SECRET")
	if err != nil {
		return SecurityConfig{}, err
	}

	return SecurityConfig{
		SessionPepper:   pepper,
		JWTSecret:       jwtSecret,
		AccessTokenTTL:  getDurationOrDefault("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getDurationOrDefault("REFRESH_TOKEN_TTL", 7*24*time.Hour),
	}, nil
}

func loadHTTP() (HTTPConfig, error) {
	timeout := getDurationOrDefault("HTTP_TIMEOUT", 10*time.Second)
	origins := getStringOrDefault("CORS_ORIGINS", "http://localhost:3000")
	corsOrigins := strings.Split(origins, ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}
	return HTTPConfig{
		Timeout:     timeout,
		CORSOrigins: corsOrigins,
	}, nil
}

func loadRedis() (RedisConfig, error) {
	host, err := getString("REDIS_HOST")
	if err != nil {
		return RedisConfig{}, err
	}

	port, err := getInt("REDIS_PORT")
	if err != nil {
		return RedisConfig{}, err
	}

	return RedisConfig{
		Host:     host,
		Port:     port,
		Password: getStringOrDefault("REDIS_PASSWORD", ""),
		DB:       getIntOrDefault("REDIS_DB", 0),
		TTL:      getDurationOrDefault("REDIS_TTL", 15*time.Minute),
	}, nil
}

func loadDatabase() (DatabaseConfig, error) {
	host, err := getString("DB_HOST")
	if err != nil {
		return DatabaseConfig{}, err
	}

	port, err := getInt("DB_PORT")
	if err != nil {
		return DatabaseConfig{}, err
	}

	user, err := getString("DB_USER")
	if err != nil {
		return DatabaseConfig{}, err
	}

	password, err := getString("DB_PASSWORD")
	if err != nil {
		return DatabaseConfig{}, err
	}

	name, err := getString("DB_NAME")
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
		// Defaults to require, not disable: an unset value against a server that
		// accepts both TLS and plaintext connects in plaintext, sending PHI and
		// DB_PASSWORD over the network with no error and no log line. A local
		// docker-compose run opts out explicitly with DB_SSLMODE=disable.
		SSLMode:      getStringOrDefault("DB_SSLMODE", "require"),
		MaxOpenConns: getIntOrDefault("DB_MAX_OPEN", 10),
		MaxIdleConns: getIntOrDefault("DB_MAX_IDLE", 5),
		MaxLifetime:  getIntOrDefault("DB_MAX_LIFETIME", 300),
		AutoMigrate:  getBoolOrDefault("DB_AUTO_MIGRATE", false),
		VerifySchema: getBoolOrDefault("DB_VERIFY_SCHEMA", true),
	}, nil
}

func loadMessaging() MessagingConfig {
	return MessagingConfig{
		Enabled: getBoolOrDefault("MESSAGING_ENABLED", false),
		Kafka: KafkaConfig{
			Brokers:         splitCSV(getStringOrDefault("KAFKA_BROKERS", "localhost:9092")),
			TopicPrefix:     getStringOrDefault("KAFKA_TOPIC_PREFIX", "auth."),
			ConsumerGroupID: getStringOrDefault("KAFKA_CONSUMER_GROUP_ID", "auth-service"),
			WriteTimeout:    getDurationOrDefault("KAFKA_WRITE_TIMEOUT", 10*time.Second),
		},
		RabbitMQ: RabbitMQConfig{
			DSN:            getStringOrDefault("RABBITMQ_DSN", "amqp://guest:guest@localhost:5672/"),
			Exchange:       getStringOrDefault("RABBITMQ_EXCHANGE", "auth.tasks"),
			RetryExchange:  getStringOrDefault("RABBITMQ_RETRY_EXCHANGE", "auth.tasks.retry"),
			DLExchange:     getStringOrDefault("RABBITMQ_DL_EXCHANGE", "auth.tasks.dlx"),
			MaxRetries:     getIntOrDefault("RABBITMQ_MAX_RETRIES", 3),
			RetryDelay:     getDurationOrDefault("RABBITMQ_RETRY_DELAY", 15*time.Second),
			PublishTimeout: getDurationOrDefault("RABBITMQ_PUBLISH_TIMEOUT", 5*time.Second),
		},
		Relay: RelayConfig{
			PollInterval:       getDurationOrDefault("RELAY_POLL_INTERVAL", 1*time.Second),
			BatchSize:          getIntOrDefault("RELAY_BATCH_SIZE", 50),
			ReclaimAfter:       getDurationOrDefault("RELAY_RECLAIM_AFTER", 2*time.Minute),
			DefaultEventBroker: getStringOrDefault("RELAY_DEFAULT_EVENT_BROKER", "event"),
			DefaultTaskBroker:  getStringOrDefault("RELAY_DEFAULT_TASK_BROKER", "task"),
			EventRoutes: map[string]string{
				"auth.user.created.v1":     "event",
				"auth.user.locked.v1":      "event",
				"auth.session.created.v1":  "event",
				"auth.session.revoked.v1":  "event",
				"auth.password.changed.v1": "event",
			},
			TaskRoutes: map[string]string{
				"auth.send-welcome-email.v1":  "task",
				"auth.send-verification.v1":   "task",
				"auth.sync-analytics-user.v1": "task",
			},
		},
	}
}

func loadTelemetry() TelemetryConfig {
	return TelemetryConfig{
		Enabled:          getBoolOrDefault("OTEL_ENABLED", false),
		ServiceName:      getStringOrDefault("OTEL_SERVICE_NAME", "medilog-api"),
		ServiceVersion:   getStringOrDefault("OTEL_SERVICE_VERSION", "0.1.0"),
		ExporterEndpoint: getStringOrDefault("OTEL_EXPORTER_ENDPOINT", "localhost:4317"),
		ExporterInsecure: getBoolOrDefault("OTEL_EXPORTER_INSECURE", true),
	}
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getDurationOrDefault(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getString(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func getStringOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getInt(key string) (int, error) {
	v, err := getString(key)
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}

	return i, nil
}

func getPositiveInt(key string) (int, error) {
	value, err := getInt(key)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func getIntOrDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getBoolOrDefault(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}

	switch v {
	case "true", "1", "yes", "y", "on":
		return true
	case "false", "0", "no", "n", "off":
		return false
	default:
		return def
	}
}

func loadApp() AppConfig {
	env := getStringOrDefault("APP_ENV", "production")

	isLive := true
	if env == "local" || env == "development" || env == "test" || env == "dev" {
		isLive = false
	}

	return AppConfig{
		IsLive: isLive,
	}
}
