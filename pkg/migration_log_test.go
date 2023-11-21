package pkg

import (
	"os"
	"strings"
	"testing"
)

const LOG_DIR = "migrations_test"
const LOG_FILE = ".log"

func cleanFiles() {
	os.RemoveAll(LOG_DIR)
}

func createLogFile(lines []string) error {
	return os.WriteFile(LOG_DIR+string(os.PathSeparator)+LOG_FILE, []byte(strings.Join(lines, "\n")), 0644)
}

// Name() returns the name of the log file without the path
func TestNameReturnsNameOfLogFile(t *testing.T) {
	migrationLog := MigrationLog{
		filePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	name := migrationLog.Name()

	if name != LOG_FILE {
		t.Fatalf("expected %s, got %s", LOG_FILE, name)
	}
}

// Count() returns the correct number of migrations
func TestCountReturnsTheCorrectNumberOfMigrations(t *testing.T) {
	defer cleanFiles()

	cases := [][]string{
		{"a", "b", "c"},
		{"a"},
		{},
	}

	os.Mkdir(LOG_DIR, 0755)

	for _, lines := range cases {
		err := createLogFile(lines)

		if err != nil {
			t.Fatal(err)
		}

		migrationLog := MigrationLog{
			filePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
		}

		err = migrationLog.Load()

		if err != nil {
			t.Fatal(err)
		}

		if migrationLog.Count() != len(lines) {
			t.Fatalf("Expected count %d, got %d", len(lines), migrationLog.Count())
		}
	}
}

// Contains() returns true if the given migration exists in the log
func TestContainsReturnsTheCorrectResult(t *testing.T) {
	defer cleanFiles()

	type testCase struct {
		migrations []string
		search     string
		expected   bool
	}

	cases := []testCase{
		{
			migrations: []string{"a", "b", "c"},
			search:     "a",
			expected:   true,
		},
		{
			migrations: []string{},
			search:     "a",
			expected:   false,
		},
		{
			migrations: []string{"migration A", "migration B"},
			search:     "migration",
			expected:   false,
		},
		{
			migrations: []string{"migration A", "migration B"},
			search:     "migration B",
			expected:   true,
		},
	}

	os.Mkdir(LOG_DIR, 0755)

	for _, testCase := range cases {
		err := createLogFile(testCase.migrations)

		if err != nil {
			t.Fatal(err)
		}

		migrationLog := MigrationLog{
			filePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
		}

		err = migrationLog.Load()

		if err != nil {
			t.Fatal(err)
		}

		if migrationLog.Contains(testCase.search) != testCase.expected {
			t.Fatalf("Expected count %t, got %t", testCase.expected, migrationLog.Contains(testCase.search))
		}
	}
}

// Add() will not update the migrations array on error
func TestAddWillNotAddMigrationsToArrayOnError(t *testing.T) {
	// Don't create the log directory, this will cause an error when writing to the log file

	migrationLog := MigrationLog{
		filePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	err := migrationLog.Add("test")

	if err == nil {
		t.Fatal("expecting error got nil")
	}

	if len(migrationLog.migrations) != 0 {
		t.Fatalf("Expected 0 migrations, got %d", len(migrationLog.migrations))
	}
}

// Add() will update the array of migrations and file
func TestAddWillUpdateTheArrayOfMigrationsAndFile(t *testing.T) {
	defer cleanFiles()

	err := os.Mkdir(LOG_DIR, 0755)

	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(LOG_DIR+string(os.PathSeparator)+LOG_FILE, []byte{}, 0644)

	if err != nil {
		t.Fatal(err)
	}

	migrationLog := MigrationLog{
		filePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	err = migrationLog.Add("test")

	if err != nil {
		t.Fatal(err)
	}

	if len(migrationLog.migrations) != 1 {
		t.Fatalf("Expected 1 migrations, got %d", len(migrationLog.migrations))
	}

	if migrationLog.migrations[0] != "test" {
		t.Fatalf("Expected 'test', got %s", migrationLog.migrations[0])
	}
}

// Pop() will not update the migrations array on error
func TestPopWillNotUpdateTheMigationsArrayOnError(t *testing.T) {
	defer cleanFiles()

	err := os.Mkdir(LOG_DIR, 0755)

	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(LOG_DIR+string(os.PathSeparator)+LOG_FILE, []byte{}, 0644)

	if err != nil {
		t.Fatal(err)
	}

	migrationLog := MigrationLog{
		filePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	migrationLog.Add("test")

	// Remove the file to trigger error on pop
	os.Remove(LOG_DIR + string(os.PathSeparator) + LOG_FILE)

	_, err = migrationLog.Pop()

	if err == nil {
		t.Fatal("expecting error got nil")
	}

	if len(migrationLog.migrations) != 1 {
		t.Fatalf("Expected 0 migrations, got %d", len(migrationLog.migrations))
	}
}

// Pop() will update the array of migrations and file
func TestPopWillUpdateTheMigationsArrayAndFile(t *testing.T) {
	defer cleanFiles()

	err := os.Mkdir(LOG_DIR, 0755)

	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(LOG_DIR+string(os.PathSeparator)+LOG_FILE, []byte{}, 0644)

	if err != nil {
		t.Fatal(err)
	}

	migrationLog := MigrationLog{
		filePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	migrationLog.Add("test")

	migration, err := migrationLog.Pop()

	if err != nil {
		t.Fatal(err)
	}

	if migration != "test" {
		t.Fatalf("expected 'test', got %s\n", migration)
	}

	if len(migrationLog.migrations) != 0 {
		t.Fatalf("Expected 0 migrations, got %d", len(migrationLog.migrations))
	}
}

// InitMigrationLog should create log file if missing
func TestInitShouldCreateLogFileIfMissing(t *testing.T) {
	defer cleanFiles()

	if _, err := os.Stat(LOG_DIR); !os.IsNotExist(err) {
		t.Fatalf("Directory %s already exists:%v", LOG_DIR, err)
	}

	filePath := LOG_DIR + string(os.PathSeparator) + LOG_FILE

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("File %s already exists", filePath)
	}

	_, err := InitMigrationLog(filePath)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal(err)
	}
}
