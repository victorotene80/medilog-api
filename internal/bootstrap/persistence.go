package bootstrap

import (
	"fmt"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"reflect"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/repository"
	outboxContracts "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/outbox"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/gormmetrics"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"github.com/victorotene80/medilog-api/internal/shared/config"
	applogging "github.com/victorotene80/medilog-api/internal/shared/logging"
	"github.com/victorotene80/medilog-api/migrations"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
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
	ReminderRepo            repository.ReminderRepository
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
	EventPublisher          appContracts.MessagePublisher
	RegisteredMedicineRepo  repository.RegisteredMedicineRepository
	SupportTicketRepo       repository.SupportTicketRepository
	SupportMessageRepo      repository.SupportMessageRepository
	SupportAttachmentRepo   repository.SupportAttachmentRepository
	FeedbackRepo            repository.FeedbackRepository
	NotificationRepo        repository.NotificationRepository
}

func initializePersistence(cfg *config.Config, logger *zap.Logger) (*Persistence, error) {
	db, err := newDatabase(cfg.Database, logger)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
		return nil, fmt.Errorf("db init failed: %w", err)
	}

	logger.Info("database connected",
		zap.String("addr", fmt.Sprintf("%s:%d", cfg.Database.Host, cfg.Database.Port)),
		zap.String("db", cfg.Database.Name),
	)

	if err := syncSchema(db, cfg.Database, logger); err != nil {
		return nil, fmt.Errorf("schema sync failed: %w", err)
	}

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

	// Built here rather than in app.go because the aggregate repositories need
	// it at construction: they drain their aggregate's domain events into the
	// outbox on the same transaction as the write. With messaging disabled this
	// is a no-op, so no rows accumulate for a relay that is not running.
	eventPublisher := newEventPublisher(outboxRepo, cfg.Messaging.Enabled, logger)

	return &Persistence{
		DB:                      db,
		EventPublisher:          eventPublisher,
		UserRepo:                userRepo,
		UserAggregateRepo:       persistence.NewUserAggregateRepository(db, eventPublisher),
		AuthProviderRepo:        authProviderRepo,
		RefreshTokenRepo:        persistence.NewRefreshTokenRepository(db),
		OTPRepo:                 persistence.NewOTPCodeRepository(db),
		CountryRepo:             persistence.NewCountryRepository(db),
		UserProfileRepo:         persistence.NewUserProfileRepository(db),
		ReminderRepo:            persistence.NewReminderRepository(db),
		FunFactRepo:             persistence.NewFunFactRepository(db),
		AllergyRepo:             persistence.NewAllergyRepository(db),
		UserAllergyRepo:         persistence.NewUserAllergyRepository(db),
		EmergencyContactRepo:    persistence.NewEmergencyContactRepository(db),
		VisitRepo:               persistence.NewVisitRepository(db),
		MedicationRepo:          persistence.NewMedicationRepository(db, eventPublisher),
		MedicationTimeRepo:      persistence.NewMedicationTimeRepository(db),
		MedicationAdherenceRepo: persistence.NewMedicationAdherenceLogRepository(db),
		DrugScanRepo:            persistence.NewDrugScanRepository(db),
		DashboardRepo:           persistence.NewDashboardRepository(db),
		AIConversationRepo:      persistence.NewAIConversationRepository(db, eventPublisher),
		AIMessageRepo:           persistence.NewAIMessageRepository(db),
		RegisteredMedicineRepo:  persistence.NewRegisteredMedicineRepository(db),
		OutboxRepo:              outboxRepo,
		SupportTicketRepo:       persistence.NewSupportTicketRepository(db, eventPublisher),
		SupportMessageRepo:      persistence.NewSupportMessageRepository(db),
		SupportAttachmentRepo:   persistence.NewSupportAttachmentRepository(db),
		FeedbackRepo:            persistence.NewFeedbackRepository(db),
		NotificationRepo:        persistence.NewNotificationRepository(db),
	}, nil
}

func newDatabase(cfg config.DatabaseConfig, logger *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: applogging.NewGormLogger(logger, applogging.GormConfig{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  gormlog.Warn,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return nil, err
	}

	if err := db.Use(gormmetrics.Plugin{}); err != nil {
		return nil, fmt.Errorf("gorm metrics plugin: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

// syncSchema keeps the database schema in line with the GORM models.
//
//   - AutoMigrate: creates/migrates tables when DB_AUTO_MIGRATE is enabled.
//   - VerifySchema: at startup, fails fast on any drift between the migration
//     files, the database, and the models.
//
// Verification refuses to boot rather than logging a warning. A missing column
// is invisible until the first request that writes it, and then it surfaces as
// whatever the calling handler makes of a database error — a missing
// user_profiles.timezone once turned every login into a 401, because the login
// handler saves the user aggregate and maps any failure to invalid credentials.
// Startup is the only place where that drift is still cheap to see.
func syncSchema(db *gorm.DB, cfg config.DatabaseConfig, logger *zap.Logger) error {
	allModels := models.All()

	if cfg.AutoMigrate {
		if err := db.AutoMigrate(allModels...); err != nil {
			return fmt.Errorf("auto-migrate: %w", err)
		}
		logger.Info("auto-migrate applied", zap.Int("tables", len(allModels)))
	}

	if !cfg.VerifySchema {
		return nil
	}

	// AutoMigrate builds the schema from the models, so schema_migrations is
	// legitimately behind (or absent) in that mode — the migration files are
	// simply not the source of truth here. Verifying the version anyway would
	// make DB_AUTO_MIGRATE=true unbootable.
	if !cfg.AutoMigrate {
		if err := verifyMigrationVersion(db, logger); err != nil {
			return err
		}
	}

	return verifySchemaShape(db, allModels, logger)
}

// verifyMigrationVersion compares the version recorded in schema_migrations
// against the newest migration embedded in this build.
func verifyMigrationVersion(db *gorm.DB, logger *zap.Logger) error {
	expected, err := migrations.LatestVersion()
	if err != nil {
		return fmt.Errorf("schema verification: %w", err)
	}
	if expected == 0 {
		return nil
	}

	if !db.Migrator().HasTable("schema_migrations") {
		return fmt.Errorf(
			"schema verification: schema_migrations table is missing, so no migration has ever been recorded "+
				"(migrations go up to %d) — run `go run ./cmd/migrate up`, "+
				"or set DB_AUTO_MIGRATE=true to build the schema from the models instead",
			expected,
		)
	}

	var row struct {
		Version uint64
		Dirty   bool
	}

	// A LIMIT keeps this correct even against the multi-row layouts other
	// migration tools use; golang-migrate itself stores exactly one row.
	if err := db.Raw("SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1").
		Scan(&row).Error; err != nil {
		return fmt.Errorf("schema verification: read schema_migrations: %w", err)
	}

	if row.Dirty {
		return fmt.Errorf(
			"schema verification: schema_migrations is dirty at version %d — a migration failed partway. "+
				"Inspect the database, finish or undo that migration by hand, then run "+
				"`go run ./cmd/migrate force %d` to clear the flag",
			row.Version, row.Version,
		)
	}

	applied := uint(row.Version)

	if applied > expected {
		// The database is ahead: usually an older binary rolled out against a
		// migrated database. The models cannot know about columns added later,
		// so this is not fatal on its own — the shape check still runs.
		logger.Warn("schema verification: database is ahead of this build",
			zap.Uint("applied", applied),
			zap.Uint("expected", expected),
			zap.String("hint", "this binary predates the latest migration; deploy the matching build"),
		)
		return nil
	}

	if applied < expected {
		return fmt.Errorf(
			"schema verification: database is at migration version %d but this build expects %d — "+
				"run `go run ./cmd/migrate up`. If the schema was originally created by "+
				"DB_AUTO_MIGRATE and its early migrations were never recorded, run "+
				"`go run ./cmd/migrate force <version already applied>` first so the "+
				"already-satisfied migrations are not replayed",
			applied, expected,
		)
	}

	logger.Info("schema verification: migrations current", zap.Uint("version", applied))
	return nil
}

// verifySchemaShape checks that every table and every column the models expect
// actually exists. Extra columns in the database are ignored: a column the
// models no longer map is harmless, and failing on one would make a
// drop-column migration impossible to deploy without downtime.
func verifySchemaShape(db *gorm.DB, allModels []any, logger *zap.Logger) error {
	// The whole live schema in one round trip. Asking GORM's migrator per model
	// would be ~30 queries, and its ColumnTypes reuses the caller's statement in
	// a way that misfires on a plain *gorm.DB.
	var rows []struct {
		TableName  string
		ColumnName string
	}
	if err := db.Raw(
		`SELECT table_name, column_name
		   FROM information_schema.columns
		  WHERE table_schema = current_schema()`,
	).Scan(&rows).Error; err != nil {
		return fmt.Errorf("schema verification: read information_schema: %w", err)
	}

	live := make(map[string]map[string]struct{})
	for _, r := range rows {
		if live[r.TableName] == nil {
			live[r.TableName] = make(map[string]struct{})
		}
		live[r.TableName][r.ColumnName] = struct{}{}
	}

	var (
		missingTables  []string
		missingColumns []string
		checkedColumns int
	)

	for _, m := range allModels {
		// Parsing the model yields the columns GORM will actually read and
		// write, which is what matters — struct fields tagged `gorm:"-"` never
		// reach a query and are correctly absent from DBNames.
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(m); err != nil {
			return fmt.Errorf("schema verification: parse model %s: %w",
				reflect.TypeOf(m).Elem().Name(), err)
		}

		actual, ok := live[stmt.Table]
		if !ok {
			missingTables = append(missingTables, stmt.Table)
			continue
		}

		for _, dbName := range stmt.Schema.DBNames {
			checkedColumns++
			if _, ok := actual[dbName]; !ok {
				missingColumns = append(missingColumns, stmt.Table+"."+dbName)
			}
		}
	}

	if len(missingTables) > 0 || len(missingColumns) > 0 {
		return fmt.Errorf(
			"schema verification failed — missing tables %v, missing columns %v; "+
				"run `go run ./cmd/migrate up`, or set DB_AUTO_MIGRATE=true to build "+
				"the schema from the models",
			missingTables, missingColumns,
		)
	}

	logger.Info("schema verification passed",
		zap.Int("tables", len(allModels)),
		zap.Int("columns", checkedColumns),
	)

	return nil
}
