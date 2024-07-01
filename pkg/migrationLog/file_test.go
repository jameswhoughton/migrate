package migrationLog

import (
	"bufio"
	"errors"
	"os"
	"strconv"
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

func readLog(fileName string) ([]Migration, error) {
	logFile, err := os.Open(LOG_DIR + string(os.PathSeparator) + fileName)

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(logFile)
	var migrations []Migration

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

		migrations = append(migrations, Migration{
			Step: step,
			Name: parts[1],
		})
	}

	return migrations, nil
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

		migrationLog := FileDriver{
			FilePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
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

		migrationLog := FileDriver{
			FilePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
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

	migrationLog := FileDriver{
		FilePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	err := migrationLog.Add(Migration{"test", 0})

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

	migrationLog := FileDriver{
		FilePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	err = migrationLog.Add(Migration{"testA", 0})

	if err != nil {
		t.Fatal(err)
	}

	err = migrationLog.Add(Migration{"testB", 0})

	if err != nil {
		t.Fatal(err)
	}

	if len(migrationLog.migrations) != 2 {
		t.Fatalf("Expected 2 migrations, got %d", len(migrationLog.migrations))
	}

	if migrationLog.migrations[0].Name != "testA" {
		t.Fatalf("Expected 'testA', got %s", migrationLog.migrations[0].Name)
	}

	if migrationLog.migrations[1].Name != "testB" {
		t.Fatalf("Expected 'testB', got %s", migrationLog.migrations[1].Name)
	}

	migrationsFromFile, err := readLog(LOG_FILE)

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

	migrationLog := FileDriver{
		FilePath: LOG_DIR + string(os.PathSeparator) + LOG_FILE,
	}

	migrationLog.Add(Migration{"test", 0})

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

	FilePath := LOG_DIR + string(os.PathSeparator) + LOG_FILE

	migrationLog := FileDriver{
		FilePath: FilePath,
	}

	err = migrationLog.Init()

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

// InitFileDriver should create log file if missing
func TestInitShouldCreateLogFileIfMissing(t *testing.T) {
	defer cleanFiles()

	if _, err := os.Stat(LOG_DIR); !os.IsNotExist(err) {
		t.Fatalf("Directory %s already exists:%v", LOG_DIR, err)
	}

	filePath := LOG_DIR + string(os.PathSeparator) + LOG_FILE

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("File %s already exists", filePath)
	}

	migrationLog := FileDriver{
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

	migrationLog := FileDriver{
		FilePath: filePath,
	}

	err = migrationLog.Init()

	if err != nil {
		t.Fatal(err)
	}

	actual := migrationLog.LastStep()

	if expected != actual {
		t.Fatalf("Expected %d got %d", expected, actual)
	}
}
