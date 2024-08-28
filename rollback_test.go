package migrate

import (
	"database/sql"
	"os"
	"testing"
	"testing/fstest"

	_ "github.com/mattn/go-sqlite3"
)

// Rollback() should run up migrations in reverse order
func TestRollbackShouldRunDownMigrationsInReverseOrder(t *testing.T) {
	db, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")

	log := newTestLog()

	testFs := fstest.MapFS{
		"1_migration_up.sql":   {Data: []byte("")},
		"1_migration_down.sql": {Data: []byte("INSERT INTO users VALUES (1, 'james')")},
		"2_migration_up.sql":   {Data: []byte("")},
		"2_migration_down.sql": {Data: []byte("CREATE TABLE users (ID INT PRIMARY KEY, name VARCHAR(100))")},
	}

	err := Migrate(db, testFs, &log)

	if err != nil {
		t.Fatal(err)
	}

	err = Rollback(db, testFs, &log)

	if err != nil {
		t.Fatal(err)
	}

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

// migrate and rollback step correctly
func TestMigrateandRollbackStepCorrectly(t *testing.T) {
	db, _ := sql.Open("sqlite3", "test.db")
	defer os.Remove("test.db")

	log := newTestLog()

	testFs := fstest.MapFS{
		"1_migrationA_up.sql":   {Data: []byte("")},
		"1_migrationA_down.sql": {Data: []byte("")},
		// "2_migrationB_up.sql":   {Data: []byte("")},
		// "2_migrationB_down.sql": {Data: []byte("")},
	}

	migrations := log.store

	if len(migrations) != 0 {
		t.Errorf("Log file should be empty, found %d migrations\n", len(migrations))
	}

	Migrate(db, testFs, &log)

	migrations = log.store

	if len(migrations) != 1 {
		t.Fatalf("Log file should contain 1 migration, found %d migrations\n", len(migrations))
	}

	if migrations[0].Step != 1 {
		t.Errorf("Expected migration to have step of 1, found %d", migrations[0].Step)
	}

	testFs["2_migrationB_up.sql"] = &fstest.MapFile{Data: []byte("")}
	testFs["2_migrationB_down.sql"] = &fstest.MapFile{Data: []byte("")}

	Migrate(db, testFs, &log)

	migrations = log.store

	if len(migrations) != 2 {
		t.Fatalf("Log file should contain 2 migrations, found %d migrations\n", len(migrations))
	}

	if migrations[1].Step != 2 {
		t.Errorf("Expected migration to have step of 2, found %d", migrations[1].Step)
	}

	err := Rollback(db, testFs, &log)

	if err != nil {
		t.Fatal(err)
	}

	migrations = log.store

	if len(migrations) != 1 {
		t.Errorf("Log file should contain 1 migration, found %d migrations\n", len(migrations))
	}

	Rollback(db, testFs, &log)

	migrations = log.store

	if len(migrations) != 0 {
		t.Errorf("Log should now be empty, found %d migrations\n", len(migrations))
	}
}
