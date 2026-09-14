package bootstrap

import (
	"fmt"
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
) (*messaging.CommandBus, appContracts.AuthService, error) {

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
		return nil, nil, fmt.Errorf("failed to create password hasher: %w", err)
	}

	sessionHasher, err := domainServices.NewSessionKeyHasher(cfg.Security.SessionPepper)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session hasher: %w", err)
	}

	tokenGen := infraServices.NewJWTGenerator(cfg.Security.JWTSecret)

	auditLogger := audit.NewAuditLogRepository(p.DB)

	sessionPolicy := policy.DefaultSessionPolicy()
	passwordPolicy := policy.DefaultPasswordPolicy()
	lockPolicy := policy.DefaultAccountLockPolicy()

	passwordService := domainServices.NewPasswordService(passwordPolicy)
	lockService := domainServices.NewAccountLockService(lockPolicy)

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
	// Lets a handler commit a state change and the outbox row describing it in
	// one transaction.
	txManager := audit.NewTransactionManager(p.DB)

	// Single owner of "end every live session for this user": revoking the
	// refresh-token rows and bumping the session version must not drift apart.
	sessionInvalidator := appServices.NewSessionInvalidator(
		p.RefreshTokenRepo,
		sessionVersionCache,
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
		sessionVersionCache,
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
		txManager,
		passwordHasher,
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
		txManager,
		clock,
	)

	requestOTPHandler := appHandlers.NewRequestOTPHandler(
		p.OTPRepo,
		p.UserRepo,
		otpService,
		ext.SMSSender,
		clock,
		cfg.App.IsLive,
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
		cfg.App.IsLive,
	)

	resetPasswordHandler := appHandlers.NewResetPasswordHandler(
		p.UserAggregateRepo,
		p.OTPRepo,
		sessionInvalidator,
		passwordHasher,
		passwordService,
		otpService,
		auditLogger,
		clock,
		cfg.App.IsLive,
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
		sessionInvalidator,
		auditLogger,
		clock,
	)

	refreshSessionHandler := appHandlers.NewRefreshSessionHandler(sessionSvc)

	deleteAccountHandler := appHandlers.NewDeleteAccountHandler(
		p.UserAggregateRepo,
		p.OTPRepo,
		sessionInvalidator,
		otpService,
		auditLogger,
		clock,
	)

	loginViaOTPHandler := appHandlers.NewLoginViaOTPHandler(
		p.UserAggregateRepo,
		p.OTPRepo,
		otpService,
		sessionSvc,
		sessionVersionCache,
		auditLogger,
		eventPublisher,
		clock,
	)

	getUserHandler := appHandlers.NewGetUserHandler(p.UserAggregateRepo)

	createEmergencyContactHandler := appHandlers.NewEmergencyContactHandler(
		p.UserAggregateRepo,
		p.EmergencyContactRepo,
		txManager,
		clock,
	)

	listEmergencyContactsHandler := appHandlers.NewListEmergencyContactsHandler(
		p.EmergencyContactRepo,
	)

	updateEmergencyContactHandler := appHandlers.NewUpdateEmergencyContactHandler(
		p.EmergencyContactRepo,
		clock,
	)

	deleteEmergencyContactHandler := appHandlers.NewDeleteEmergencyContactHandler(
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

	updateUserAllergyHandler := appHandlers.NewUpdateUserAllergyHandler(
		p.UserAllergyRepo,
		clock,
	)

	generateDueRemindersHandler := appHandlers.NewGenerateDueRemindersHandler(
		p.ReminderRepo,
		p.NotificationRepo,
		logger,
		clock,
	)

	getAIQuotaHandler := appHandlers.NewGetAIQuotaHandler(
		p.UserProfileRepo,
		clock,
	)

	updateUserHandler := appHandlers.NewUpdateUserHandler(
		p.UserAggregateRepo,
		clock,
	)

	getNotificationPreferencesHandler := appHandlers.NewGetNotificationPreferencesHandler(
		p.UserProfileRepo,
	)

	updateNotificationPreferencesHandler := appHandlers.NewUpdateNotificationPreferencesHandler(
		p.UserProfileRepo,
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
		p.MedicationRepo, p.MedicationAdherenceRepo, txManager)

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
		p.UserProfileRepo,
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

	createSupportTicketHandler := appHandlers.NewCreateSupportTicketHandler(
		p.SupportTicketRepo,
		clock,
	)

	getSupportTicketHandler := appHandlers.NewGetSupportTicketHandler(
		p.SupportTicketRepo,
		clock,
	)

	listSupportTicketsHandler := appHandlers.NewListSupportTicketsHandler(
		p.SupportTicketRepo,
		clock,
	)

	addSupportMessageHandler := appHandlers.NewAddSupportMessageHandler(
		p.SupportTicketRepo,
		clock,
	)

	submitFeedbackHandler := appHandlers.NewSubmitFeedbackHandler(
		p.FeedbackRepo,
		clock,
	)

	getFeedbackHandler := appHandlers.NewGetFeedbackHandler(
		p.FeedbackRepo,
		clock,
	)

	listNotificationsHandler := appHandlers.NewListNotificationsHandler(
		p.NotificationRepo,
		clock,
	)

	getNotificationHandler := appHandlers.NewGetNotificationHandler(
		p.NotificationRepo,
		clock,
	)

	markNotificationReadHandler := appHandlers.NewMarkNotificationReadHandler(
		p.NotificationRepo,
		clock,
	)

	markAllNotificationsReadHandler := appHandlers.NewMarkAllNotificationsReadHandler(
		p.NotificationRepo,
		clock,
	)

	listAuditLogsHandler := appHandlers.NewListAuditLogsHandler(
		auditLogger,
	)

	messaging.MustRegister(bus, verifDrugScanHandler)
	messaging.MustRegister(bus, listDrugScansHandler)
	messaging.MustRegister(bus, getDrugScanHandler)
	messaging.MustRegister(bus, getDashboardHandler)
	messaging.MustRegister(bus, createSupportTicketHandler)
	messaging.MustRegister(bus, getSupportTicketHandler)
	messaging.MustRegister(bus, listSupportTicketsHandler)
	messaging.MustRegister(bus, addSupportMessageHandler)
	messaging.MustRegister(bus, submitFeedbackHandler)
	messaging.MustRegister(bus, getFeedbackHandler)
	messaging.MustRegister(bus, listNotificationsHandler)
	messaging.MustRegister(bus, getNotificationHandler)
	messaging.MustRegister(bus, markNotificationReadHandler)
	messaging.MustRegister(bus, markAllNotificationsReadHandler)
	messaging.MustRegister(bus, listAuditLogsHandler)
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
	updateAIConversationHandler := appHandlers.NewUpdateAIConversationHandler(
		p.AIConversationRepo,
		clock,
	)

	messaging.MustRegister(bus, createAIConversationHandler)
	messaging.MustRegister(bus, listAIConversationsHandler)
	messaging.MustRegister(bus, getAIConversationHandler)
	messaging.MustRegister(bus, sendAIMessageHandler)
	messaging.MustRegister(bus, archiveAIConversationHandler)
	messaging.MustRegister(bus, updateAIConversationHandler)

	messaging.MustRegister(bus, createAllergyHandler)
	messaging.MustRegister(bus, updateAllergyHandler)
	messaging.MustRegister(bus, deleteAllergyHandler)
	messaging.MustRegister(bus, createFunFactHandler)
	messaging.MustRegister(bus, updateFunFactHandler)
	messaging.MustRegister(bus, deleteFunFactHandler)
	messaging.MustRegister(bus, deleteUserAllergyHandler)
	messaging.MustRegister(bus, updateUserAllergyHandler)
	messaging.MustRegister(bus, getAIQuotaHandler)
	messaging.MustRegister(bus, generateDueRemindersHandler)
	messaging.MustRegister(bus, updateUserHandler)
	messaging.MustRegister(bus, getNotificationPreferencesHandler)
	messaging.MustRegister(bus, updateNotificationPreferencesHandler)
	messaging.MustRegister(bus, createUserAllergiesHandler)
	messaging.MustRegister(bus, createEmergencyContactHandler)
	messaging.MustRegister(bus, listEmergencyContactsHandler)
	messaging.MustRegister(bus, updateEmergencyContactHandler)
	messaging.MustRegister(bus, deleteEmergencyContactHandler)

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
	messaging.MustRegister(bus, refreshSessionHandler)
	messaging.MustRegister(bus, deleteAccountHandler)
	messaging.MustRegister(bus, loginViaOTPHandler)

	messaging.MustRegister(bus, getUserHandler)

	messaging.MustRegister[appQuery.GetCountriesQuery, []dto.CountryDTO](bus, listCountriesHandler)
	messaging.MustRegister[appQuery.GetAllergiesQuery, []dto.AllergyDTO](bus, getAllergiesHandler)
	messaging.MustRegister[appQuery.ListFunFactsQuery, []dto.FunFactDTO](bus, listFunFactsHandler)
	messaging.MustRegister[appQuery.GetFunFactQuery, *dto.FunFactDTO](bus, getFunFactHandler)
	messaging.MustRegister[appQuery.ListUserAllergiesQuery, []dto.UserAllergyDTO](bus, listUserAllergiesHandler)
	return bus, authSvc, nil
}
