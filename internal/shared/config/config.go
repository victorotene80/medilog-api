package config

import "time"

type Config struct {
	App       AppConfig
	Security  SecurityConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Messaging MessagingConfig
	HTTP      HTTPConfig
	Google    GoogleConfig
	AI        AIConfig
	Claude    ClaudeConfig
	OTP       OTPConfig
	Telemetry TelemetryConfig
	Twilio    TwilioConfig
	BulkSms   BulkSmsConfig
	SMS       SMSConfig
	Telnyx    TelnyxConfig
	Scheduler SchedulerConfig
}

type TelnyxConfig struct {
	APIKey             string
	FromNumber         string
	MessagingProfileID string
	MessagesURL        string
}

type SMSConfig struct {
	Enabled         bool
	MaxBulkFailures int
}

type AIConfig struct {
	Enabled               bool
	Provider              string
	ContextWindowTokens   int
	MaxContextMessages    int
	SummaryTokenThreshold int
}

type ClaudeConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	APIVersion  string
	WorkspaceID string
}

type BulkSmsConfig struct {
	ProdBaseURL        string
	TestBaseURL        string
	SendMessagePath    string
	CheckBalancePath   string
	DeliveryReportPath string
	APIToken           string
	LegacyToken        string
	Sender             string
}

type TwilioConfig struct {
	AccountSID         string
	AuthToken          string
	WhatsAppFromNumber string
	// Use either FromNumber or MessagingServiceSID.
	// MessagingServiceSID is better for production.
	FromNumber          string
	MessagingServiceSID string
	MessagesURL         string
	BaseURL             string
}

type MessagingConfig struct {
	Enabled  bool
	Kafka    KafkaConfig
	RabbitMQ RabbitMQConfig
	Relay    RelayConfig
}

type GoogleConfig struct {
	ClientID     string
	TokenInfoURL string
	UserInfoURL  string
	OAuthBaseURL string
}

type KafkaConfig struct {
	Brokers         []string
	TopicPrefix     string
	ConsumerGroupID string
	WriteTimeout    time.Duration
}

type HTTPConfig struct {
	Timeout     time.Duration
	CORSOrigins []string
}

type RabbitMQConfig struct {
	DSN            string
	Exchange       string
	RetryExchange  string
	DLExchange     string
	MaxRetries     int
	RetryDelay     time.Duration
	PublishTimeout time.Duration
}

type RelayConfig struct {
	PollInterval       time.Duration
	BatchSize          int
	ReclaimAfter       time.Duration
	DefaultEventBroker string
	DefaultTaskBroker  string
	EventRoutes        map[string]string
	TaskRoutes         map[string]string
}

// SchedulerConfig drives the reminder ticker. See
// internal/infrastructure/scheduler for what each value means.
type SchedulerConfig struct {
	Enabled       bool
	PollInterval  time.Duration
	CatchupWindow time.Duration
	LeadTime      time.Duration
	BatchSize     int
}

type SecurityConfig struct {
	SessionPepper   string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int
	AutoMigrate  bool
	VerifySchema bool
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	TTL      time.Duration
}

type OTPConfig struct {
	Length     int
	TTLMins    int
	BcryptCost int
}

type TelemetryConfig struct {
	Enabled          bool
	ServiceName      string
	ServiceVersion   string
	ExporterEndpoint string
	ExporterInsecure bool
}

type AppConfig struct {
	IsLive bool
}
