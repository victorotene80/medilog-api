package bootstrap

import (
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/victorotene80/medilog-api/internal/application/messaging"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/infrastructure/validation"

	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest"
	restHandler "github.com/victorotene80/medilog-api/internal/interfaces/http/rest/handler"
	appmw "github.com/victorotene80/medilog-api/internal/interfaces/middleware"
	"github.com/victorotene80/medilog-api/internal/shared/config"
)

func initializeHTTP(
	commandBus *messaging.CommandBus,
	logger *zap.Logger,
	authSvc appContracts.AuthService,
	redisClient *redis.Client,
	db *gorm.DB,
	telemetryEnabled bool,
	cfg *config.Config,
) *rest.Router {
	validate := validation.NewPlaygroundValidator()

	//  Auth handlers
	authHandler := restHandler.NewAuthHandler(commandBus, validate)
	otpHandler := restHandler.NewOTPHandler(commandBus, validate)

	//  User handlers
	userHandler := restHandler.NewUserHandler(commandBus, validate)

	//  Health record handlers
	allergyHandler := restHandler.NewAllergyHandler(commandBus, validate)
	funFactHandler := restHandler.NewFunFactHandler(commandBus, validate)
	userAllergyHandler := restHandler.NewUserAllergyHandler(commandBus, validate)
	emergencyContactHandler := restHandler.NewEmergencyContactHandler(commandBus, validate)
	medicationHandler := restHandler.NewMedicationHandler(commandBus, validate)
	visitHandler := restHandler.NewVisitHandler(commandBus, validate)
	aiHandler := restHandler.NewAIHandler(commandBus, validate)
	dashboardHandler := restHandler.NewDashboardHandler(commandBus, validate)

	//  Reference data handlers ─
	referenceHandler := restHandler.NewReferenceHandler(commandBus, validate)
	scanHandler := restHandler.NewScanHandler(commandBus, validate)

	//  Support handlers
	supportTicketHandler := restHandler.NewSupportTicketHandler(commandBus, validate)

	//  Feedback handlers
	feedbackHandler := restHandler.NewFeedbackHandler(commandBus, validate)

	//  Notification handlers
	notificationHandler := restHandler.NewNotificationHandler(commandBus, validate)

	//  Audit log handlers
	auditLogHandler := restHandler.NewAuditLogHandler(commandBus, validate)

	//  Health handlers
	healthHandler := restHandler.NewHealthHandler(db, redisClient)

	authMiddleware := appmw.NewAuthMiddleware(authSvc, logger)
	adminMiddleware := appmw.NewAdminMiddleware(authSvc, logger)
	rateLimiter := appmw.NewRateLimiter(redisClient, logger)

	var telemetryMiddleware *appmw.TelemetryMiddleware
	if telemetryEnabled {
		tm, err := appmw.NewTelemetryMiddleware()
		if err != nil {
			logger.Fatal("failed to create telemetry middleware", zap.Error(err))
		}
		telemetryMiddleware = tm
	}

	router := rest.NewRouter(
		logger,
		authMiddleware,
		adminMiddleware,
		rateLimiter,
		telemetryMiddleware,
		cfg.HTTP.CORSOrigins,
		authHandler,
		otpHandler,
		userHandler,
		userAllergyHandler,
		emergencyContactHandler,
		medicationHandler,
		visitHandler,
		aiHandler,
		referenceHandler,
		allergyHandler,
		scanHandler,
		funFactHandler,
		dashboardHandler,
		supportTicketHandler,
		feedbackHandler,
		notificationHandler,
		auditLogHandler,
		healthHandler,
		cfg.App.IsLive,
	)

	router.Setup()
	return router
}
