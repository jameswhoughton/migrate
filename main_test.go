package migrate

import (
	"bufio"
	"database/sql"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/jameswhoughton/migrate/pkg/migrationLog"
	_ "github.com/mattn/go-sqlite3"
)

const MIGRATION_DIR = "migrations_test"

func cleanFiles() {
	os.RemoveAll(MIGRATION_DIR)
}

func readLog(fileName string) ([]migrationLog.Migration, error) {
	logFile, err := os.Open(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(logFile)
	var migrations []migrationLog.Migration

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

		migrations = append(migrations, migrationLog.Migration{
			Step: step,
			Name: parts[1],
		})
	}

	return migrations, nil
}

// creates migrations dir if doesn't exist
func TestCreatesMigrationDirectoryIfMissing(t *testing.T) {
	defer cleanFiles()

	Create(MIGRATION_DIR, "test")

	if _, err := os.Stat(MIGRATION_DIR); os.IsNotExist(err) {
		t.Fatal("migrations directory not found")
	}
}

// Create() returns the names of the migration pair
func TestCreateReturnsNamesOfMigrationPair(t *testing.T) {
	defer cleanFiles()

	migration := Create(MIGRATION_DIR, "test")

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

	Create(MIGRATION_DIR, "test")

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

	Create(MIGRATION_DIR, "test 123 !bg**,TEST")

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

// Migrate() should return error if the query fails to execute
func TestReturnsErrorIfNQueryFails(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	os.Mkdir(MIGRATION_DIR, 0755)
	migrationLog, err := migrationLog.Init(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	migration := Create(MIGRATION_DIR, "migration")

	os.WriteFile(MIGRATION_DIR+string(os.PathSeparator)+migration.up, []byte("I am not a valid query"), os.ModeAppend)

	if err != nil {
		t.Fatal(err)
	}

	err = Migrate(conn, MIGRATION_DIR, migrationLog)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	got, isCorrectType := err.(ErrorQuery)

	if !isCorrectType {
		t.Fatalf("Expected ErrorQuery error, got %s", got)
	}
}

// Migrate() should log migrations in .log
func TestMigrateShouldLogMigrations(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	migrationName := "create_user_table"

	migrationPair := Create(MIGRATION_DIR, migrationName)
	migrationLog, err := migrationLog.Init(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if err != nil {
		t.Fatal(err)
	}

	err = Migrate(conn, MIGRATION_DIR, migrationLog)

	if err != nil {
		t.Fatal(err)
	}

	migrations, err := readLog(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if err != nil {
		t.Fatal(err)
	}

	if len(migrations) != 1 {
		t.Errorf("Expected 1 migration got %d\n", len(migrations))
	}

	if strings.HasPrefix(migrations[0].Name, migrationPair.Name()) {
		t.Errorf("Expected text `%s` got `%s`", migrationName, migrations[0].Name)
	}
}

// Migrate() should run up migrations in order
func TestMigrateShouldRunUpMigrationsInOrder(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	migrationA := Create(MIGRATION_DIR, "migration")
	migrationB := Create(MIGRATION_DIR, "migration")

	os.WriteFile(MIGRATION_DIR+string(os.PathSeparator)+migrationA.up, []byte("CREATE TABLE users (ID INT PRIMARY KEY, name VARCHAR(100))"), os.ModeAppend)
	os.WriteFile(MIGRATION_DIR+string(os.PathSeparator)+migrationB.down, []byte("INSERT INTO users VALUES ('james')"), os.ModeAppend)

	migrationLog, err := migrationLog.Init(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if err != nil {
		t.Fatal(err)
	}

	Migrate(conn, MIGRATION_DIR, migrationLog)

	query, err := conn.Query("SELECT name FROM users")

	if err != nil {
		t.Errorf("Error when querying users %v\n", err)
	}
	defer query.Close()
	var name string

	for query.Next() {
		query.Scan(&name)

		if name != "james" {
			t.Errorf("Expected 'james' got %s\n", name)
		}
	}
}

// Rollback() should run up migrations in reverse order
func TestRollbackShouldRunDownMigrationsInReverseOrder(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	migrationLog, err := migrationLog.Init(MIGRATION_DIR + string(os.PathSeparator) + ".log")
	migrationA := Create(MIGRATION_DIR, "migration")
	migrationB := Create(MIGRATION_DIR, "migration")

	os.WriteFile(MIGRATION_DIR+string(os.PathSeparator)+migrationB.down, []byte("CREATE TABLE users (ID INT PRIMARY KEY, name VARCHAR(100))"), os.ModeAppend)
	os.WriteFile(MIGRATION_DIR+string(os.PathSeparator)+migrationA.down, []byte("INSERT INTO users VALUES ('james')"), os.ModeAppend)

	if err != nil {
		t.Fatal(err)
	}

	Migrate(conn, MIGRATION_DIR, migrationLog)

	Rollback(conn, MIGRATION_DIR, migrationLog)

	query, err := conn.Query("SELECT name FROM users")

	if err != nil {
		t.Errorf("Error when querying users %v\n", err)
	}
	defer query.Close()
	var name string

	for query.Next() {
		query.Scan(&name)

		if name != "james" {
			t.Errorf("Expected 'james' got %s\n", name)
		}
	}

}

// migrate and rollback step correctly
func TestMigrateandRollbackStepCorrectly(t *testing.T) {
	conn, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")
	defer cleanFiles()

	migrationLog, err := migrationLog.Init(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if err != nil {
		t.Fatal(err)
	}

	Create(MIGRATION_DIR, "migrationA")

	migrations, err := readLog(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if err != nil {
		t.Fatal(err)
	}

	if len(migrations) != 0 {
		t.Errorf("Log file should be empty, found %d migrations\n", len(migrations))
	}

	Migrate(conn, MIGRATION_DIR, migrationLog)

	migrations, _ = readLog(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if len(migrations) != 1 {
		t.Fatalf("Log file should contain 1 migration, found %d migrations\n", len(migrations))
	}

	if migrations[0].Step != 1 {
		t.Errorf("Expected migration to have step of 1, found %d", migrations[0].Step)
	}

	Create(MIGRATION_DIR, "migrationB")

	Migrate(conn, MIGRATION_DIR, migrationLog)

	migrations, _ = readLog(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if len(migrations) != 2 {
		t.Fatalf("Log file should contain 2 migrations, found %d migrations\n", len(migrations))
	}

	if migrations[1].Step != 2 {
		t.Errorf("Expected migration to have step of 2, found %d", migrations[1].Step)
	}

	err = Rollback(conn, MIGRATION_DIR, migrationLog)

	if err != nil {
		t.Fatal(err)
	}

	migrations, _ = readLog(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if len(migrations) != 1 {
		t.Errorf("Log file should contain 1 migration, found %d migrations\n", len(migrations))
	}

	Rollback(conn, MIGRATION_DIR, migrationLog)

	migrations, _ = readLog(MIGRATION_DIR + string(os.PathSeparator) + ".log")

	if len(migrations) != 0 {
		t.Errorf("Log should now be empty, found %d migrations\n", len(migrations))
	}
}
