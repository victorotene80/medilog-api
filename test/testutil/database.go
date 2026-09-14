package testutil

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	t.Cleanup(func() {
		sqlDB.Close()
	})

	return db
}

func CleanupTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	tables := []string{
		"audit_logs",
		"outbox_events",
		"notifications",
		"feedback",
		"support_attachments",
		"support_messages",
		"support_tickets",
		"ai_messages",
		"ai_conversations",
		"fun_facts",
		"registered_medicines",
		"drug_scans",
		"emergency_contacts",
		"user_allergies",
		"allergies",
		"medication_adherence_logs",
		"medication_times",
		"visits",
		"medications",
		"otp_codes",
		"refresh_tokens",
		"user_auth_providers",
		"user_profiles",
		"users",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			t.Logf("warning: could not truncate %s: %v", table, err)
		}
	}
}
