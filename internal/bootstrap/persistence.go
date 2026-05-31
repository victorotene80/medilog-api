package bootstrap

import (
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/repository"
	outboxContracts "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/outbox"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence"
	"github.com/victorotene80/medilog-api/internal/shared/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Persistence struct {
	DB *gorm.DB

	UserRepo                repository.UserRepository
	UserAggregateRepo       repository.UserAggregateRepository
	AuthProviderRepo        repository.UserAuthProviderRepository
	RefreshTokenRepo        repository.RefreshTokenRepository
	OTPRepo                 repository.OTPCodeRepository
	CountryRepo             repository.CountryRepository
	UserProfileRepo         repository.UserProfileRepository
	FunFactRepo             repository.FunFactRepository
	AllergyRepo             repository.AllergyRepository
	UserAllergyRepo         repository.UserAllergyRepository
	EmergencyContactRepo    repository.EmergencyContactRepository
	VisitRepo               repository.VisitRepository
	MedicationRepo          repository.MedicationRepository
	MedicationTimeRepo      repository.MedicationTimeRepository
	MedicationAdherenceRepo repository.MedicationAdherenceLogRepository
	DrugScanRepo            repository.DrugScanRepository
	DashboardRepo           repository.DashboardRepository
	AIConversationRepo      repository.AIConversationRepository
	AIMessageRepo           repository.AIMessageRepository
	OutboxRepo              outboxContracts.OutboxRepository
	RegisteredMedicineRepo  repository.RegisteredMedicineRepository
}

func initializePersistence(cfg *config.Config, logger *zap.Logger) (*Persistence, error) {
	db, err := newDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
		return nil, fmt.Errorf("db init failed: %w", err)
	}

	logger.Info("database connected",
		zap.String("addr", fmt.Sprintf("%s:%d", cfg.Database.Host, cfg.Database.Port)),
		zap.String("db", cfg.Database.Name),
	)

	userRepo, err := persistence.NewUserRepository(db)
	if err != nil {
		return nil, fmt.Errorf("user repo init failed: %w", err)
	}

	authProviderRepo, err := persistence.NewUserAuthProviderRepository(db)
	if err != nil {
		return nil, fmt.Errorf("auth provider repo init failed: %w", err)
	}

	outboxRepo, err := persistence.NewOutboxRepository(db)
	if err != nil {
		return nil, fmt.Errorf("outbox repo init failed: %w", err)
	}

	return &Persistence{
		DB:                      db,
		UserRepo:                userRepo,
		UserAggregateRepo:       persistence.NewUserAggregateRepository(db),
		AuthProviderRepo:        authProviderRepo,
		RefreshTokenRepo:        persistence.NewRefreshTokenRepository(db),
		OTPRepo:                 persistence.NewOTPCodeRepository(db),
		CountryRepo:             persistence.NewCountryRepository(db),
		UserProfileRepo:         persistence.NewUserProfileRepository(db),
		FunFactRepo:             persistence.NewFunFactRepository(db),
		AllergyRepo:             persistence.NewAllergyRepository(db),
		UserAllergyRepo:         persistence.NewUserAllergyRepository(db),
		EmergencyContactRepo:    persistence.NewEmergencyContactRepository(db),
		VisitRepo:               persistence.NewGormVisitRepository(db),
		MedicationRepo:          persistence.NewMedicationRepository(db),
		MedicationTimeRepo:      persistence.NewMedicationTimeRepository(db),
		MedicationAdherenceRepo: persistence.NewMedicationAdherenceLogRepository(db),
		DrugScanRepo:            persistence.NewDrugScanRepository(db),
		DashboardRepo:           persistence.NewDashboardRepository(db),
		AIConversationRepo:      persistence.NewAIConversationRepository(db),
		AIMessageRepo:           persistence.NewAIMessageRepository(db),
		RegisteredMedicineRepo:  persistence.NewRegisteredMedicineRepository(db),
		OutboxRepo:              outboxRepo,
	}, nil
}

func newDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	return db, nil
}
