package migrationLog

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

// Name() returns the Name of the log file without the path
func TestNameReturnsNameOfLogFile(t *testing.T) {
	migrationLog := MigrationLog{
		filePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	Name := migrationLog.Name()

	if Name != LOG_FILE {
		t.Fatalf("expected %s, got %s", LOG_FILE, Name)
	}
}

// Count() returns the correct number of migrations
func TestCountReturnsTheCorrectNumberOfMigrations(t *testing.T) {
	defer cleanFiles()

	cases := [][]string{
		{"0,a", "1,b", "1,c"},
		{"0,a"},
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
			migrations: []string{"0,a", "0,b", "0,c"},
			search:     "a",
			expected:   true,
		},
		{
			migrations: []string{},
			search:     "a",
			expected:   false,
		},
		{
			migrations: []string{"0,migration A", "1,migration B"},
			search:     "migration",
			expected:   false,
		},
		{
			migrations: []string{"0,migration A", "1,migration B"},
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

	err := migrationLog.Add("test", 0)

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

	err = migrationLog.Add("test", 0)

	if err != nil {
		t.Fatal(err)
	}

	if len(migrationLog.migrations) != 1 {
		t.Fatalf("Expected 1 migrations, got %d", len(migrationLog.migrations))
	}

	if migrationLog.migrations[0].Name != "test" {
		t.Fatalf("Expected 'test', got %s", migrationLog.migrations[0].Name)
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

	migrationLog.Add("test", 0)

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

	migrations := []string{"0,a", "1,b", "1,c"}

	err = createLogFile(migrations)

	if err != nil {
		t.Fatal(err)
	}

	filePath := LOG_DIR + string(os.PathSeparator) + LOG_FILE

	migrationLog, err := Init(filePath)

	if err != nil {
		t.Fatal(err)
	}

	migration, err := migrationLog.Pop()

	if err != nil {
		t.Fatal(err)
	}

	if migration.Name != "c" {
		t.Fatalf("Expected migration 'c' got '%s", migration.Name)
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

	_, err := Init(filePath)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal(err)
	}
}

// NextStep returns the next available step index
func TestLastStepReturnsNextAvaiableIndex(t *testing.T) {
	defer cleanFiles()

	migrations := []string{"0,a", "1,b", "1,c"}

	expected := 1

	os.Mkdir(LOG_DIR, 0755)

	err := createLogFile(migrations)

	if err != nil {
		t.Fatal(err)
	}

	filePath := LOG_DIR + string(os.PathSeparator) + LOG_FILE

	migrationLog, err := Init(filePath)

	if err != nil {
		t.Fatal(err)
	}

	actual := migrationLog.LastStep()

	if expected != actual {
		t.Fatalf("Expected %d got %d", expected, actual)
	}
}
