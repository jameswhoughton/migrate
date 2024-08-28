package migrate

import (
	"os"
	"testing"
)

const MIGRATION_DIR = "migrations"

// MakeMigration() returns the name of the migration
func TestPrefixSuffixCanBeAddedToMigration(t *testing.T) {
	defer os.RemoveAll(MIGRATION_DIR)

	os.Mkdir(MIGRATION_DIR, 0755)

	MakeMigration(MIGRATION_DIR, "test", "AAA", "")

	MakeMigration(MIGRATION_DIR, "test", "", "AAA")

	files, _ := os.ReadDir(MIGRATION_DIR)

	if len(files) != 2 {
		t.Fatalf("expected 2 migrations, got %d\n", len(files))
	}

	if files[0].Name() != "AAA_test_.sql" {
		t.Fatalf("expected name to be AAA_test_.sql, got %s\n", files[0].Name())
	}

	if files[1].Name() != "_test_AAA.sql" {
		t.Fatalf("expected name to be testAAA.sql, got %s\n", files[0].Name())
	}
}

// Any non-alphanumeric characters in the name should be replaced with an underscore
func TestNonAlphaNumericCharactersShouldBeReplacedWithUnderscore(t *testing.T) {
	defer os.RemoveAll(MIGRATION_DIR)

	os.Mkdir(MIGRATION_DIR, 0755)

	MakeMigration(MIGRATION_DIR, "test 123 !bg**,TEST", "", "")

	files, err := os.ReadDir(MIGRATION_DIR)

	if err != nil {
		t.Fatalf("could not read directory: %s (%s)\n", MIGRATION_DIR, err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d\n", len(files))
	}

	if files[0].Name() != "_test_123_bg_TEST_.sql" {
		t.Fatalf("migration doesn't match the expected format, got: %s\n", files[0].Name())
	}
}
