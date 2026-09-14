package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found or failed to load:", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := getEnvOrDefault("DB_HOST", "localhost")
		port := getEnvOrDefault("DB_PORT", "5432")
		user := getEnvOrDefault("DB_USER", "postgres")
		password := getEnvOrDefault("DB_PASSWORD", "postgres")
		name := getEnvOrDefault("DB_NAME", "medilog")
		sslmode := getEnvOrDefault("DB_SSLMODE", "disable")
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslmode)
	}

	source := "file://migrations"

	// Without Parse the flag package holds no positional args, so flag.Arg(0)
	// is always "" and every invocation silently fell through to "status" —
	// `migrate up` could never apply anything.
	flag.Parse()

	action := flag.Arg(0)
	if action == "" {
		action = "status"
	}

	m, err := migrate.New(source, dsn)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migrations rolled back successfully")

	case "status":
		version, dirty, err := m.Version()
		if err != nil {
			log.Printf("Migration status: no migrations applied yet (error: %v)", err)
		} else {
			log.Printf("Migration status: version=%d dirty=%v", version, dirty)
		}

	case "goto":
		if flag.NArg() < 2 {
			log.Fatal("Usage: migrate goto <version>")
		}
		version := flag.Arg(1)
		var v uint
		if _, err := fmt.Sscanf(version, "%d", &v); err != nil {
			log.Fatalf("Invalid version number: %s", version)
		}
		if err := m.Migrate(v); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration to version %d failed: %v", v, err)
		}
		log.Printf("Migrated to version %d", v)

	// force records a version without running its SQL. Two situations need it:
	// a failed migration left schema_migrations dirty, and a schema that was
	// built by GORM AutoMigrate before the migration files existed — there the
	// early migrations are already satisfied but unrecorded, and replaying a
	// non-idempotent one (000003 adds columns without IF NOT EXISTS) fails.
	case "force":
		if flag.NArg() < 2 {
			log.Fatal("Usage: migrate force <version>")
		}
		var v int
		if _, err := fmt.Sscanf(flag.Arg(1), "%d", &v); err != nil {
			log.Fatalf("Invalid version number: %s", flag.Arg(1))
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("Force to version %d failed: %v", v, err)
		}
		log.Printf("Forced version to %d (no SQL was run)", v)

	case "create":
		if flag.NArg() < 2 {
			log.Fatal("Usage: migrate create <name>")
		}
		log.Printf("Create migration files manually in migrations/ directory")

	default:
		log.Fatalf("Unknown action: %s. Available: up, down, status, goto, force, create", action)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
