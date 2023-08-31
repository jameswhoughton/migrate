package main

import (
	"database/sql"
	"os"
	"regexp"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type testOS struct {
	files       []string
	directories []string
}

func (t *testOS) Stat(name string) (os.FileInfo, error) {
	for _, filename := range t.files {
		if filename == name {
			return nil, nil
		}
	}

	for _, dirname := range t.directories {
		if dirname == name {
			return nil, nil
		}
	}

	return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
}

func (*testOS) IsNotExist(err error) bool { return os.IsNotExist(err) }

func (t *testOS) Mkdir(name string, perm os.FileMode) error {
	t.directories = append(t.directories, name)

	return nil
}

func (t *testOS) WriteFile(name string, data []byte, perm os.FileMode) error {
	t.files = append(t.files, name)

	return nil
}

// creates migrations dir if doesn't exist
func TestCreatesMigrationDirectoryIfMissing(t *testing.T) {
	tOS := &testOS{}

	createMigration(tOS, "migrations", "test")

	if len(tOS.directories) != 1 || tOS.directories[0] != "migrations" {
		t.Fatal("migrations directory not found in: " + strings.Join(tOS.directories, ", "))
	}
}

// creates migration pair
func TestCreatesAMigrationPair(t *testing.T) {
	tOS := &testOS{}

	createMigration(tOS, "migrations", "test")

	if len(tOS.files) != 2 {
		t.Fatalf("Expected 2 files, got %d\n", len(tOS.files))
	}

	upRegexp, _ := regexp.Compile("[0-9]+_test_up.sql")
	downRegexp, _ := regexp.Compile("[0-9]+_test_down.sql")

	if !upRegexp.Match([]byte(tOS.files[0])) {
		t.Fatalf("Up migration doesn't match the expected format, got: %s\n", tOS.files[0])
	}

	if !downRegexp.Match([]byte(tOS.files[1])) {
		t.Fatalf("Down migration doesn't match the expected format, got: %s\n", tOS.files[0])
	}
}

// Any non-alphanumeric characters in the name should be replaced with an underscore
func TestNonAlphaNumericCharactersShouldBeReplacedWithUnderscore(t *testing.T) {
	tOS := &testOS{}

	createMigration(tOS, "migrations", "test 123 !bg**,TEST")

	if len(tOS.files) != 2 {
		t.Fatalf("Expected 2 files found %d \n", len(tOS.files))
	}

	upRegexp, _ := regexp.Compile("[0-9]+_test_123_bg_TEST_up.sql")
	downRegexp, _ := regexp.Compile("[0-9]+_test_123_bg_TEST_down.sql")

	if !upRegexp.Match([]byte(tOS.files[0])) {
		t.Fatalf("Up migration doesn't match the expected format, got: %s\n", tOS.files[0])
	}

	if !downRegexp.Match([]byte(tOS.files[1])) {
		t.Fatalf("Down migration doesn't match the expected format, got: %s\n", tOS.files[0])
	}
}

// runMigrations() should create .log if missing
func TestCreateslogFileIfMissing(t *testing.T) {
	tOS := &testOS{}
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")

	runMigrations(tOS, conn, "migrations")

	if len(tOS.files) != 1 || !strings.HasSuffix(tOS.files[0], ".log") {
		t.Fatal(".log file not found: " + strings.Join(tOS.files, ", "))
	}
}

// runMigrations() should return error if there are no migrations to run
func TestReturnsErrorIfNoMigrationsToRun(t *testing.T) {
	tOS := &testOS{}
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")

	err := runMigrations(tOS, conn, "migrations")

	if err == nil {
		t.Fatalf("Expected error %s, got nil", ERROR_NO_MIGRATIONS)
	}

	if err == ERROR_NO_MIGRATIONS {
		t.Fatalf("Expected error %s, got %s", ERROR_NO_MIGRATIONS, err)
	}

}
