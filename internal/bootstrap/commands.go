package bootstrap

import (
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	appHandlers "github.com/victorotene80/medilog-api/internal/application/handlers"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	msgmw "github.com/victorotene80/medilog-api/internal/application/messaging/middleware"
	appQuery "github.com/victorotene80/medilog-api/internal/application/query"
	appServices "github.com/victorotene80/medilog-api/internal/application/services"
	domainServices "github.com/victorotene80/medilog-api/internal/domain/services"
	"github.com/victorotene80/medilog-api/internal/domain/services/policy"

	audit "github.com/victorotene80/medilog-api/internal/infrastructure/persistence"
	cacheInfra "github.com/victorotene80/medilog-api/internal/infrastructure/persistence/cache"
	infraServices "github.com/victorotene80/medilog-api/internal/infrastructure/services"

	"github.com/victorotene80/medilog-api/internal/shared/config"
	"golang.org/x/crypto/bcrypt"
)

func initializeCommands(
	p *Persistence,
	ext ExternalServices,
	eventPublisher appContracts.MessagePublisher,
	cfg *config.Config,
	redisClient *redis.Client,
	logger *zap.Logger,
) (*messaging.CommandBus, appContracts.AuthService) {

	bus := messaging.NewCommandBus()
	clock := func() time.Time { return time.Now().UTC() }

	if logger != nil {
		msgmw.AttachLogging(bus, logger)
	}

	passwordHasher, err := infraServices.NewBcryptPasswordHasher(
		cfg.Security.SessionPepper,
		bcrypt.DefaultCost,
	)
	if err != nil {
		panic(err)
	}

	sessionHasher, err := domainServices.NewSessionKeyHasher(cfg.Security.SessionPepper)
	if err != nil {
		panic(err)
	}

	tokenGen := infraServices.NewJWTGenerator(cfg.Security.JWTSecret)

	auditLogger := audit.NewAuditLogRepository(p.DB)

	sessionPolicy := policy.DefaultSessionPolicy()
	passwordPolicy := policy.DefaultPasswordPolicy()
	lockPolicy := policy.DefaultAccountLockPolicy()

	passwordService := domainServices.NewPasswordService(passwordPolicy)
	lockService := domainServices.NewAccountLockService(lockPolicy)

	//otpPolicy := policy()
	otpConfig := cfg.OTP
	otpService := domainServices.NewOTPService(
		otpConfig.Length,
		time.Duration(otpConfig.TTLMins)*time.Minute,
		otpConfig.BcryptCost,
		time.Now,
	)

	sessionCache := cacheInfra.NewRedisCache[string, appContracts.CachedToken](
		redisClient,
		"session:",
		sessionPolicy.MaxDuration,
	)
	sessionVersionCache := cacheInfra.NewSessionCache(
		redisClient,
		sessionPolicy.MaxDuration,
	)
	countriesCache := cacheInfra.NewRedisCache[string, []dto.CountryDTO](
		redisClient,
		"reference:countries:",
		time.Hour,
	)

	authSvc := appServices.NewAuthService(
		tokenGen,
		p.RefreshTokenRepo,
		p.UserRepo,
		sessionPolicy,
		sessionCache,
		clock,
		sessionVersionCache,
	)

	sessionSvc := appServices.NewSessionService(
		p.RefreshTokenRepo,
		tokenGen,
		sessionHasher,
		sessionPolicy,
		sessionCache,
		clock,
	)

	loginHandler := appHandlers.NewLoginHandler(
		p.UserAggregateRepo,
		passwordHasher,
		sessionSvc,
		lockService,
		eventPublisher,
		clock,
	)

	registerHandler := appHandlers.NewRegisterHandler(
		p.UserAggregateRepo,
		passwordHasher,
		sessionSvc,
		eventPublisher,
		clock,
		passwordService,
	)

	googleLoginHandler := appHandlers.NewGoogleAuthHandler(
		p.UserRepo,
		p.UserAggregateRepo,
		p.AuthProviderRepo,
		ext.GoogleAuth,
		sessionSvc,
		eventPublisher,
		clock,
	)

	requestOTPHandler := appHandlers.NewRequestOTPHandler(
		p.OTPRepo,
		p.UserRepo,
		otpService,
		ext.SMSSender,
		clock,
	)
	verifyOTPHandler := appHandlers.NewVerifyOTPHandler(
		p.OTPRepo,
		otpService,
		clock,
	)

	verifyOnboardingOTPHandler := appHandlers.NewVerifyOnboardingOTPHandler(
		p.UserRepo,
		p.OTPRepo,
		otpService,
		sessionSvc,
		clock,
	)

	forgotPasswordHandler := appHandlers.NewForgotPasswordHandler(
		p.UserRepo,
		p.OTPRepo,
		otpService,
		ext.SMSSender,
		auditLogger,
		clock,
	)

	resetPasswordHandler := appHandlers.NewResetPasswordHandler(
		p.UserAggregateRepo,
		p.OTPRepo,
		p.RefreshTokenRepo,
		passwordHasher,
		passwordService,
		otpService,
		auditLogger,
		clock,
	)

	logoutHandler := appHandlers.NewLogoutHandler(
		p.RefreshTokenRepo,
		sessionCache,
		auditLogger,
		clock,
	)

	changePasswordHandler := appHandlers.NewChangePasswordHandler(
		p.UserAggregateRepo,
		passwordHasher,
		passwordService,
		p.RefreshTokenRepo,
		sessionCache,
		auditLogger,
		sessionVersionCache,
		clock,
	)

	getUserHandler := appHandlers.NewGetUserHandler(p.UserAggregateRepo)

	createEmergencyContactHandler := appHandlers.NewEmergencyContactHandler(
		p.UserRepo,
		p.EmergencyContactRepo,
		clock,
	)

	listCountriesHandler := appHandlers.NewGetCountriesHandler(
		p.CountryRepo,
		countriesCache,
	)
	getAllergiesHandler := appHandlers.NewGetAllergiesHandler(p.AllergyRepo)

	createAllergyHandler := appHandlers.NewAllergyCreationHandler(
		p.AllergyRepo,
		eventPublisher,
		clock,
	)

	updateAllergyHandler := appHandlers.NewUpdateAllergyHandler(
		p.AllergyRepo,
	)

	deleteAllergyHandler := appHandlers.NewDeleteAllergyHandler(
		p.AllergyRepo,
	)

	createFunFactHandler := appHandlers.NewCreateFunFactHandler(
		p.FunFactRepo,
	)
	updateFunFactHandler := appHandlers.NewUpdateFunFactHandler(
		p.FunFactRepo,
	)
	deleteFunFactHandler := appHandlers.NewDeleteFunFactHandler(
		p.FunFactRepo,
	)
	listFunFactsHandler := appHandlers.NewListFunFactsHandler(
		p.FunFactRepo,
	)
	getFunFactHandler := appHandlers.NewGetFunFactHandler(
		p.FunFactRepo,
	)

	createUserAllergiesHandler := appHandlers.NewCreateUserAllergiesHandler(
		p.UserAllergyRepo,
		p.AllergyRepo,
		clock,
	)

	listUserAllergiesHandler := appHandlers.NewListUserAllergiesHandler(
		p.UserAllergyRepo,
	)

	deleteUserAllergyHandler := appHandlers.NewDeleteUserAllergyHandler(
		p.UserAllergyRepo,
	)

	createMedicationHandler := appHandlers.NewCreateMedicationHandler(
		p.MedicationRepo,
	)

	updateMedicationHandler := appHandlers.NewUpdateMedicationHandler(
		p.MedicationRepo,
		p.MedicationTimeRepo,
	)

	completeMedicationHandler := appHandlers.NewCompleteMedicationHandler(
		p.MedicationRepo,
	)

	deleteMedicationHandler := appHandlers.NewDeleteMedicationHandler(
		p.MedicationRepo,
	)

	listMedicationHandler := appHandlers.NewListMedicationsHandler(
		p.MedicationRepo)

	getMedicationHandler := appHandlers.NewGetMedicationHandler(
		p.MedicationRepo)

	logMedicationAdherenceHandler := appHandlers.NewLogMedicationAdherenceHandler(
		p.MedicationRepo, p.MedicationAdherenceRepo)

	aiSvc := appServices.NewAIService(ext.AIModel, clock)
	aiContextBuilder := appServices.NewAIContextBuilder(
		p.UserProfileRepo,
		p.UserAllergyRepo,
		p.MedicationRepo,
		p.VisitRepo,
		p.DrugScanRepo,
		clock,
	)

	createAIConversationHandler := appHandlers.NewCreateAIConversationHandler(
		p.AIConversationRepo,
		p.MedicationRepo,
		p.VisitRepo,
		clock,
	)
	listAIConversationsHandler := appHandlers.NewListAIConversationsHandler(
		p.AIConversationRepo,
	)
	getAIConversationHandler := appHandlers.NewGetAIConversationHandler(
		p.AIConversationRepo,
	)
	sendAIMessageHandler := appHandlers.NewSendAIMessageHandler(
		p.AIConversationRepo,
		p.AIMessageRepo,
		aiContextBuilder,
		aiSvc,
		clock,
		cfg.AI.ContextWindowTokens,
		cfg.AI.MaxContextMessages,
		cfg.AI.SummaryTokenThreshold,
	)
	archiveAIConversationHandler := appHandlers.NewArchiveAIConversationHandler(
		p.AIConversationRepo,
		clock,
	)

	createVisitHandler := appHandlers.NewCreateVisitHandler(
		p.VisitRepo)

	updateVisitHandler := appHandlers.NewUpdateVisitHandler(
		p.VisitRepo)

	deleteVisitHandler := appHandlers.NewDeleteVisitHandler(
		p.VisitRepo)

	listVisitHandler := appHandlers.NewListVisitsHandler(
		p.VisitRepo)

	getVisitHandler := appHandlers.NewGetVisitHandler(
		p.VisitRepo)

	verifDrugScanHandler := appHandlers.NewVerifyDrugScanHandler(
		p.DrugScanRepo,
		p.RegisteredMedicineRepo,
		clock,
	)

	listDrugScansHandler := appHandlers.NewListDrugScansHandler(
		p.DrugScanRepo,
	)

	getDrugScanHandler := appHandlers.NewGetDrugScanHandler(
		p.DrugScanRepo,
	)

	getDashboardHandler := appHandlers.NewGetDashboardHandler(
		p.DashboardRepo,
		clock,
	)

	messaging.MustRegister(bus, verifDrugScanHandler)
	messaging.MustRegister(bus, listDrugScansHandler)
	messaging.MustRegister(bus, getDrugScanHandler)
	messaging.MustRegister(bus, getDashboardHandler)
	messaging.MustRegister(bus, createMedicationHandler)
	messaging.MustRegister(bus, updateMedicationHandler)
	messaging.MustRegister(bus, completeMedicationHandler)
	messaging.MustRegister(bus, deleteMedicationHandler)
	messaging.MustRegister(bus, listMedicationHandler)
	messaging.MustRegister(bus, getMedicationHandler)
	messaging.MustRegister(bus, logMedicationAdherenceHandler)
	messaging.MustRegister(bus, createVisitHandler)
	messaging.MustRegister(bus, updateVisitHandler)
	messaging.MustRegister(bus, deleteVisitHandler)
	messaging.MustRegister(bus, listVisitHandler)
	messaging.MustRegister(bus, getVisitHandler)
	messaging.MustRegister(bus, createAIConversationHandler)
	messaging.MustRegister(bus, listAIConversationsHandler)
	messaging.MustRegister(bus, getAIConversationHandler)
	messaging.MustRegister(bus, sendAIMessageHandler)
	messaging.MustRegister(bus, archiveAIConversationHandler)

	messaging.MustRegister(bus, createAllergyHandler)
	messaging.MustRegister(bus, updateAllergyHandler)
	messaging.MustRegister(bus, deleteAllergyHandler)
	messaging.MustRegister(bus, createFunFactHandler)
	messaging.MustRegister(bus, updateFunFactHandler)
	messaging.MustRegister(bus, deleteFunFactHandler)
	messaging.MustRegister(bus, deleteUserAllergyHandler)
	messaging.MustRegister(bus, createUserAllergiesHandler)
	messaging.MustRegister(bus, createEmergencyContactHandler)

	messaging.MustRegister(bus, loginHandler)
	messaging.MustRegister(bus, registerHandler)
	messaging.MustRegister(bus, googleLoginHandler)
	messaging.MustRegister(bus, requestOTPHandler)
	messaging.MustRegister(bus, verifyOTPHandler)
	messaging.MustRegister(bus, verifyOnboardingOTPHandler)
	messaging.MustRegister(bus, forgotPasswordHandler)
	messaging.MustRegister(bus, resetPasswordHandler)
	messaging.MustRegister(bus, logoutHandler)
	messaging.MustRegister(bus, changePasswordHandler)

	messaging.MustRegister(bus, getUserHandler)

	messaging.MustRegister[appQuery.GetCountriesQuery, []dto.CountryDTO](bus, listCountriesHandler)
	messaging.MustRegister[appQuery.GetAllergiesQuery, []dto.AllergyDTO](bus, getAllergiesHandler)
	messaging.MustRegister[appQuery.ListFunFactsQuery, []dto.FunFactDTO](bus, listFunFactsHandler)
	messaging.MustRegister[appQuery.GetFunFactQuery, *dto.FunFactDTO](bus, getFunFactHandler)
	messaging.MustRegister[appQuery.ListUserAllergiesQuery, []dto.UserAllergyDTO](bus, listUserAllergiesHandler)
	return bus, authSvc
}
