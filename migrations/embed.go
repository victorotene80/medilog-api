// Package migrations embeds the SQL migration files so the schema version the
// binary expects travels with the binary itself.
//
// The startup drift check (DB_VERIFY_SCHEMA) needs to know the newest migration
// that exists in this build. Reading the migrations/ directory off disk would
// work in development and fail in the container, where only the compiled binary
// ships — and a check that silently no-ops in production is worse than no check
// at all. Embedding makes the expected version a build-time constant.
//
// golang-migrate's file source skips any filename that does not match
// `<version>_<name>.<up|down>.<ext>`, so this .go file is invisible to it.
package migrations

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4/source"
)

//go:embed *.sql
var FS embed.FS

// LatestVersion returns the highest migration version embedded in this build,
// or 0 when there are no migration files at all.
func LatestVersion() (uint, error) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		return 0, fmt.Errorf("read embedded migrations: %w", err)
	}

	var latest uint
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		// Same parser golang-migrate uses, so this can never disagree with the
		// version the migrate command would report.
		m, err := source.DefaultParse(e.Name())
		if err != nil {
			continue
		}

		if m.Version > latest {
			latest = m.Version
		}
	}

	return latest, nil
}
