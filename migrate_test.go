package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// Migrate() should return error if the query fails to execute
func TestReturnsErrorIfQueryFails(t *testing.T) {
	db, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")

	log := newTestLog()

	testFs := fstest.MapFS{
		"1_migration_up.sql": {Data: []byte("I am not a valid query")},
	}

	err := Migrate(db, testFs, &log)

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
	db, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")

	migrationName := "create_user_table"

	testFs := fstest.MapFS{
		"1_" + migrationName + "_up.sql": {Data: []byte("")},
	}

	log := newTestLog()

	err := Migrate(db, testFs, &log)

	if err != nil {
		t.Fatal(err)
	}

	migrations := log.store

	if err != nil {
		t.Fatal(err)
	}

	fmt.Print(migrations)

	if len(migrations) != 1 {
		t.Errorf("Expected 1 migration got %d\n", len(migrations))
	}

	if strings.HasPrefix(migrations[0].Name, "1_"+migrationName+"_up.sql") {
		t.Errorf("Expected text `%s` got `%s`", migrationName, migrations[0].Name)
	}
}

// Migrate() should run up migrations in order
func TestMigrateShouldRunUpMigrationsInNameOrder(t *testing.T) {
	db, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")

	testFs := fstest.MapFS{
		"2_migration_up.sql": {Data: []byte("INSERT INTO users VALUES ('james')")},
		"1_migration_up.sql": {Data: []byte("CREATE TABLE users (ID INT PRIMARY KEY, name VARCHAR(100))")},
	}
	log := newTestLog()

	Migrate(db, testFs, &log)

	query, err := db.Query("SELECT name FROM users")

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
