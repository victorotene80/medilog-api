package bootstrap

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/victorotene80/medilog-api/internal/application/messaging"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/infrastructure/validation"

	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest"
	restHandler "github.com/victorotene80/medilog-api/internal/interfaces/http/rest/handler"
	appmw "github.com/victorotene80/medilog-api/internal/interfaces/middleware"
)

func initializeHTTP(
	commandBus *messaging.CommandBus,
	logger *zap.Logger,
	authSvc appContracts.AuthService,
	redisClient *redis.Client,

) http.Handler {
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

	authMiddleware := appmw.NewAuthMiddleware(authSvc, logger)
	rateLimiter := appmw.NewRateLimiter(redisClient, logger)

	// Telemetry middleware — wraps every route with a server span + metrics.
	// NewTelemetryMiddleware returns an error only when the global MeterProvider
	// has not been set, which cannot happen here because initializeTelemetry
	// runs before initializeHTTP in InitializeApp.
	telemetryMiddleware, err := appmw.NewTelemetryMiddleware()
	if err != nil {
		// Non-recoverable: the meter is misconfigured.
		log.Fatalf("failed to create telemetry middleware: %v", err)
	}

	router := rest.NewRouter(
		logger,
		authMiddleware,
		rateLimiter,
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
	)

	return telemetryMiddleware.Handle(router.Setup())
}
