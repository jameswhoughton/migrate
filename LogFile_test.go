package migrate_test

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/jameswhoughton/migrate"
)

const LOG_DIR = "migrations_test"
const LOG_FILE = ".log"

func readLog() ([]migrate.Migration, error) {
	logFile, err := os.Open(LOG_DIR + string(os.PathSeparator) + LOG_FILE)

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(logFile)
	var migrations []migrate.Migration

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(scanner.Text(), ",")

		if len(parts) != 2 {
			return nil, errors.New("log file line malformed: " + line)
		}

		step, err := strconv.Atoi(parts[0])

		if err != nil {
			return nil, errors.Join(errors.New("Invalid step: "+parts[0]), err)
		}

		migrations = append(migrations, migrate.Migration{
			Step: step,
			Name: parts[1],
		})
	}

	return migrations, nil
}

func createLogFile(lines []string) error {
	return os.WriteFile(LOG_DIR+string(os.PathSeparator)+LOG_FILE, []byte(strings.Join(lines, "\n")), 0644)
}

// Contains() returns true if the given migration exists in the log
func TestFileContainsReturnsTheCorrectResult(t *testing.T) {
	defer os.RemoveAll(LOG_DIR)

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

		migrationLog, err := migrate.NewLogFile(LOG_DIR + string(os.PathSeparator) + LOG_FILE)

		if err != nil {
			t.Fatal(err)
		}

		if migrationLog.Contains(testCase.search) != testCase.expected {
			t.Fatalf("Expected count %t, got %t", testCase.expected, migrationLog.Contains(testCase.search))
		}
	}
}

// Add() will not update the migrations array on error
func TestFileAddWillNotAddMigrationsToArrayOnError(t *testing.T) {
	defer os.RemoveAll(LOG_DIR)
	// Don't create the log directory, this will cause an error when writing to the log file

	migrationLog := migrate.LogFile{
		FilePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	err := migrationLog.Add(migrate.Migration{"test", 0})

	if err == nil {
		t.Fatal("expecting error got nil")
	}

	if len(migrationLog.Migrations) != 0 {
		t.Fatalf("Expected 0 migrations, got %d", len(migrationLog.Migrations))
	}
}

// Add() will update the array of migrations and file
func TestFileAddWillUpdateTheArrayOfMigrationsAndFile(t *testing.T) {
	defer os.RemoveAll(LOG_DIR)

	err := os.Mkdir(LOG_DIR, 0755)

	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(LOG_DIR+string(os.PathSeparator)+LOG_FILE, []byte{}, 0644)

	if err != nil {
		t.Fatal(err)
	}

	migrationLog := migrate.LogFile{
		FilePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	err = migrationLog.Add(migrate.Migration{"testA", 0})

	if err != nil {
		t.Fatal(err)
	}

	err = migrationLog.Add(migrate.Migration{"testB", 0})

	if err != nil {
		t.Fatal(err)
	}

	if len(migrationLog.Migrations) != 2 {
		t.Fatalf("Expected 2 migrations, got %d", len(migrationLog.Migrations))
	}

	if migrationLog.Migrations[0].Name != "testA" {
		t.Fatalf("Expected 'testA', got %s", migrationLog.Migrations[0].Name)
	}

	if migrationLog.Migrations[1].Name != "testB" {
		t.Fatalf("Expected 'testB', got %s", migrationLog.Migrations[1].Name)
	}

	migrationsFromFile, err := readLog()

	if err != nil {
		t.Fatal(err)
	}

	if len(migrationsFromFile) != 2 {
		t.Fatalf("Expected 2 migrations, got %d", len(migrationsFromFile))
	}

	if migrationsFromFile[0].Name != "testA" {
		t.Fatalf("Expected 'testA', got %s", migrationsFromFile[0].Name)
	}

	if migrationsFromFile[1].Name != "testB" {
		t.Fatalf("Expected 'testB', got %s", migrationsFromFile[1].Name)
	}
}

// Pop() will not update the migrations array on error
func TestFilePopWillNotUpdateTheMigationsArrayOnError(t *testing.T) {
	defer os.RemoveAll(LOG_DIR)

	err := os.Mkdir(LOG_DIR, 0755)

	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(LOG_DIR+string(os.PathSeparator)+LOG_FILE, []byte{}, 0644)

	if err != nil {
		t.Fatal(err)
	}

	migrationLog := migrate.LogFile{
		FilePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	migrationLog.Add(migrate.Migration{"test", 0})

	// Remove the file to trigger error on pop
	os.Remove(LOG_DIR + string(os.PathSeparator) + LOG_FILE)

	_, err = migrationLog.Pop()

	if err == nil {
		t.Fatal("expecting error got nil")
	}

	if len(migrationLog.Migrations) != 1 {
		t.Fatalf("Expected 0 migrations, got %d", len(migrationLog.Migrations))
	}
}

// Pop() will update the array of migrations and file
func TestFilePopWillUpdateTheMigationsArrayAndFile(t *testing.T) {
	defer os.RemoveAll(LOG_DIR)

	err := os.Mkdir(LOG_DIR, 0755)

	if err != nil {
		t.Fatal(err)
	}

	migrations := []string{"0,a", "1,b", "1,c"}

	err = createLogFile(migrations)

	if err != nil {
		t.Fatal(err)
	}

	migrationLog, err := migrate.NewLogFile(LOG_DIR + string(os.PathSeparator) + LOG_FILE)

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

// InitLogFile should create log file if missing
func TestFileInitShouldCreateLogFileIfMissing(t *testing.T) {
	defer os.RemoveAll(LOG_DIR)

	if _, err := os.Stat(LOG_DIR); !os.IsNotExist(err) {
		t.Fatalf("Directory %s already exists:%v", LOG_DIR, err)
	}

	filePath := LOG_DIR + string(os.PathSeparator) + LOG_FILE

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("File %s already exists", filePath)
	}

	migrationLog := migrate.LogFile{
		FilePath: filePath,
	}

	err := migrationLog.Init()

	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal(err)
	}
}

// NextStep returns the next available step index
func TestFileLastStepReturnsNextAvaiableIndex(t *testing.T) {
	defer os.RemoveAll(LOG_DIR)

	migrations := []string{"0,a", "1,b", "1,c"}

	expected := 1

	os.Mkdir(LOG_DIR, 0755)

	err := createLogFile(migrations)

	if err != nil {
		t.Fatal(err)
	}

	migrationLog, err := migrate.NewLogFile(LOG_DIR + string(os.PathSeparator) + LOG_FILE)

	if err != nil {
		t.Fatal(err)
	}

	actual := migrationLog.LastStep()

	if expected != actual {
		t.Fatalf("Expected %d got %d", expected, actual)
	}
}
