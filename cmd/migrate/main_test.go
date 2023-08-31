package main

import (
	"database/sql"
	"os"
	"regexp"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const MIGRATION_DIR = "migrations_test"

func cleanFiles() {
	os.RemoveAll(MIGRATION_DIR)
}

// creates migrations dir if doesn't exist
func TestCreatesMigrationDirectoryIfMissing(t *testing.T) {
	defer cleanFiles()

	createMigration("migrations", "test")

	if _, err := os.Stat(MIGRATION_DIR); os.IsNotExist(err) {
		t.Fatal("migrations directory not found")
	}
}

// creates migration pair
func TestCreatesAMigrationPair(t *testing.T) {
	defer cleanFiles()

	createMigration(MIGRATION_DIR, "test")

	files, err := os.ReadDir(MIGRATION_DIR)

	if err != nil {
		t.Fatalf("could not read directory: %s (%s)\n", MIGRATION_DIR, err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d\n", len(files))
	}

	upRegexp, _ := regexp.Compile("[0-9]+_test_up.sql")
	downRegexp, _ := regexp.Compile("[0-9]+_test_down.sql")

	if !upRegexp.Match([]byte(files[0].Name())) {
		t.Fatalf("Up migration doesn't match the expected format, got: %s\n", files[0].Name())
	}

	if !downRegexp.Match([]byte(files[1].Name())) {
		t.Fatalf("Down migration doesn't match the expected format, got: %s\n", files[1].Name())
	}
}

// Any non-alphanumeric characters in the name should be replaced with an underscore
func TestNonAlphaNumericCharactersShouldBeReplacedWithUnderscore(t *testing.T) {
	defer cleanFiles()

	createMigration(MIGRATION_DIR, "test 123 !bg**,TEST")

	files, err := os.ReadDir(MIGRATION_DIR)

	if err != nil {
		t.Fatalf("could not read directory: %s (%s)\n", MIGRATION_DIR, err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d\n", len(files))
	}

	upRegexp, _ := regexp.Compile("[0-9]+_test_123_bg_TEST_up.sql")
	downRegexp, _ := regexp.Compile("[0-9]+_test_123_bg_TEST_down.sql")

	if !upRegexp.Match([]byte(files[0].Name())) {
		t.Fatalf("Up migration doesn't match the expected format, got: %s\n", files[0].Name())
	}

	if !downRegexp.Match([]byte(files[1].Name())) {
		t.Fatalf("Down migration doesn't match the expected format, got: %s\n", files[1].Name())
	}
}

// runMigrations() should create .log if missing
func TestCreateslogFileIfMissing(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	runMigrations(conn, MIGRATION_DIR)

	if _, err := os.Stat(MIGRATION_DIR + string(os.PathSeparator) + ".log"); os.IsNotExist(err) {
		t.Fatal(".log file not found")
	}
}

// runMigrations() should return error if there are no migrations to run
func TestReturnsErrorIfNoMigrationsToRun(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	err := runMigrations(conn, "migrations")
	expected := ErrorNoMigrations{}

	if err == nil {
		t.Fatalf("Expected error %s, got nil", expected)
	}

	got, isCorrectType := err.(ErrorNoMigrations)

	if !isCorrectType {
		t.Fatalf("Expected error %s, got %s", expected, got)
	}

}
