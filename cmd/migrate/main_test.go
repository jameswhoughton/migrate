package main

import (
	"bufio"
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

	create(MIGRATION_DIR, "test")

	if _, err := os.Stat(MIGRATION_DIR); os.IsNotExist(err) {
		t.Fatal("migrations directory not found")
	}
}

// create() returns the names of the migration pair
func TestCreateReturnsNamesOfMigrationPair(t *testing.T) {
	defer cleanFiles()

	migration := create(MIGRATION_DIR, "test")

	upRegexp, _ := regexp.Compile("[0-9]+_test_up.sql")
	downRegexp, _ := regexp.Compile("[0-9]+_test_down.sql")

	if !upRegexp.Match([]byte(migration.up)) {
		t.Fatalf("Up migration doesn't match the expected format, got: %s\n", migration.up)
	}

	if !downRegexp.Match([]byte(migration.down)) {
		t.Fatalf("Down migration doesn't match the expected format, got: %s\n", migration.down)
	}
}

// creates migration pair
func TestCreatesAMigrationPair(t *testing.T) {
	defer cleanFiles()

	create(MIGRATION_DIR, "test")

	files, err := os.ReadDir(MIGRATION_DIR)

	if err != nil {
		t.Fatalf("could not read directory: %s (%s)\n", MIGRATION_DIR, err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d\n", len(files))
	}

	upRegexp, _ := regexp.Compile("[0-9]+_test_up.sql")
	downRegexp, _ := regexp.Compile("[0-9]+_test_down.sql")

	if !upRegexp.Match([]byte(files[1].Name())) {
		t.Fatalf("Up migration doesn't match the expected format, got: %s\n", files[1].Name())
	}

	if !downRegexp.Match([]byte(files[0].Name())) {
		t.Fatalf("Down migration doesn't match the expected format, got: %s\n", files[0].Name())
	}
}

// Any non-alphanumeric characters in the name should be replaced with an underscore
func TestNonAlphaNumericCharactersShouldBeReplacedWithUnderscore(t *testing.T) {
	defer cleanFiles()

	create(MIGRATION_DIR, "test 123 !bg**,TEST")

	files, err := os.ReadDir(MIGRATION_DIR)

	if err != nil {
		t.Fatalf("could not read directory: %s (%s)\n", MIGRATION_DIR, err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d\n", len(files))
	}

	upRegexp, _ := regexp.Compile("[0-9]+_test_123_bg_TEST_up.sql")
	downRegexp, _ := regexp.Compile("[0-9]+_test_123_bg_TEST_down.sql")

	if !upRegexp.Match([]byte(files[1].Name())) {
		t.Fatalf("Up migration doesn't match the expected format, got: %s\n", files[1].Name())
	}

	if !downRegexp.Match([]byte(files[0].Name())) {
		t.Fatalf("Down migration doesn't match the expected format, got: %s\n", files[0].Name())
	}
}

// migrate() should create .log if missing
func TestCreateslogFileIfMissing(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	os.Mkdir(MIGRATION_DIR, 0755)

	migrate(conn, MIGRATION_DIR)

	if _, err := os.Stat(MIGRATION_DIR + string(os.PathSeparator) + ".log"); os.IsNotExist(err) {
		t.Fatal(".log file not found")
	}
}

// migrate() should return error if there are no migrations to run
func TestReturnsErrorIfNoMigrationsToRun(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	os.Mkdir(MIGRATION_DIR, 0755)

	err := migrate(conn, MIGRATION_DIR)
	expected := ErrorNoMigrations{}

	if err == nil {
		t.Fatalf("Expected error %s, got nil", expected)
	}

	got, isCorrectType := err.(ErrorNoMigrations)

	if !isCorrectType {
		t.Fatalf("Expected error %s, got %s", expected, got)
	}
}

// migrate() should log migrations in .log
func TestMigrateShouldLogMigrations(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	migrationName := "create_user_table"

	create(MIGRATION_DIR, migrationName)

	err := migrate(conn, MIGRATION_DIR)

	if err != nil {
		t.Fatal(err)
	}

	logFile, err := os.Open(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(logFile)
	var migrations []string

	for scanner.Scan() {
		migrations = append(migrations, scanner.Text())
	}

	if len(migrations) != 1 {
		t.Errorf("Expected 1 migration got %d\n", len(migrations))
	}

	if migrations[0] != migrationName {
		t.Errorf("Expected text `%s` got `%s`", migrationName, migrations[0])
	}
}

// rollback() should error if log file missing
func TestRollbackShouldErrorIfLogFileMissing(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	err := rollback(conn, MIGRATION_DIR)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// migrate() should run up migrations in order
func TestMigrateShouldRunUpMigrationsInOrder(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	migrationA := create(MIGRATION_DIR, "migration_a")
	migrationB := create(MIGRATION_DIR, "migration_a")

	os.WriteFile(MIGRATION_DIR+string(os.PathSeparator)+migrationA.up, []byte("CREATE TABLE users (ID INT PRIMARY KEY, name VARCHAR(100))"), os.ModeAppend)
	os.WriteFile(MIGRATION_DIR+string(os.PathSeparator)+migrationB.up, []byte("INSERT INTO users VALUES ('james')"), os.ModeAppend)

	migrate(conn, MIGRATION_DIR)

	query, err := conn.Query("SELECT count(*) AS count FROM users")

	if err != nil {
		t.Errorf("Error when querying users %v\n", err)
	}

	var count int64

	query.Scan(&count)

	if count != 1 {
		t.Errorf("Expected 1 got %d\n", count)
	}

}

// migrate() should run up migrations in order
func TestRollbackShouldRunDownMigrationsInReverseOrder(t *testing.T) {

}
