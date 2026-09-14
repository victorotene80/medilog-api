package migrations

import (
	"os"
	"testing"
)

// The startup drift check fails the boot when the database is behind
// LatestVersion, so a wrong answer here either blocks every deploy or silently
// disables the check.
func TestLatestVersionMatchesFilesOnDisk(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	var onDisk int
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 7 && e.Name()[len(e.Name())-7:] == ".up.sql" {
			onDisk++
		}
	}

	got, err := LatestVersion()
	if err != nil {
		t.Fatal(err)
	}

	// Versions are sequential and start at 1, so the newest version and the
	// number of up files agree — and disagreeing means a gap or a duplicate
	// version, which golang-migrate would refuse to run anyway.
	if got != uint(onDisk) {
		t.Fatalf("LatestVersion() = %d but found %d .up.sql files; "+
			"migration versions must be sequential with no gaps or duplicates", got, onDisk)
	}
}

func TestLatestVersionIgnoresNonMigrationFiles(t *testing.T) {
	// embed.go itself lives in this directory; if it were counted, or if
	// golang-migrate's parser rejected it, this package would misreport.
	if _, err := FS.ReadFile("embed.go"); err == nil {
		t.Fatal("embed.go must not be part of the embedded migration set")
	}

	got, err := LatestVersion()
	if err != nil {
		t.Fatal(err)
	}
	if got == 0 {
		t.Fatal("LatestVersion() = 0, expected the embedded migrations to be found")
	}
}
