package migrate_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/jameswhoughton/migrate"
	_ "github.com/mattn/go-sqlite3"
)

func sqliteDb() (*sql.DB, func(), error) {
	db, err := sql.Open("sqlite3", "test.db")

	if err != nil {
		return nil, nil, err
	}

	return db, func() {
		os.Remove("test.db")
	}, nil

}

func TestNewLogSQLiteCreatesMigrationsTable(t *testing.T) {
	db, tearDown, err := sqliteDb()
	defer tearDown()

	if err != nil {
		t.Fatal(err)
	}

	_, err = migrate.NewLogSQLite(db)

	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow("SELECT COUNT(*) FROM sqlite_schema WHERE type='table' AND name='migrations';")

	var count int

	row.Scan(&count)

	if count != 1 {
		t.Errorf("Migration table missing")
	}
}

// Contains() returns true if the given migration exists in the log
func TestSQLiteContainsReturnsTheCorrectResult(t *testing.T) {
	type testCase struct {
		name       string
		migrations []migrate.Migration
		search     string
		expected   bool
	}

	cases := []testCase{
		{
			name: "search in list",
			migrations: []migrate.Migration{
				{"a", 0},
				{"b", 0},
				{"c", 0},
			},
			search:   "a",
			expected: true,
		},
		{
			name:       "empty migrations",
			migrations: []migrate.Migration{},
			search:     "a",
			expected:   false,
		},
		{
			name: "partial search",
			migrations: []migrate.Migration{
				{"migration A", 0},
				{"migration B", 0},
			},
			search:   "migration",
			expected: false,
		},
		{
			name: "different steps",
			migrations: []migrate.Migration{
				{"migration A", 0},
				{"migration B", 1},
			},
			search:   "migration B",
			expected: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			db, tearDown, err := sqliteDb()
			defer tearDown()

			if err != nil {
				t.Fatal(err)
			}

			migrationLog, err := migrate.NewLogSQLite(db)

			if err != nil {
				t.Fatal(err)
			}

			for _, migration := range testCase.migrations {
				db.Exec("INSERT INTO migrations (name, step) VALUES (?, ?);", migration.Name, migration.Step)
			}

			if migrationLog.Contains(testCase.search) != testCase.expected {
				t.Fatalf("Expected count %t, got %t", testCase.expected, migrationLog.Contains(testCase.search))
			}

		})
	}
}

func TestSQLiteAddInsertsMigrationIntoTable(t *testing.T) {
	db, tearDown, err := sqliteDb()
	defer tearDown()

	if err != nil {
		t.Fatal(err)
	}

	log, err := migrate.NewLogSQLite(db)

	if err != nil {
		t.Fatal(err)
	}

	countQuery := db.QueryRow("SELECT COUNT(*) FROM migrations;")

	var count int

	countQuery.Scan(&count)

	if count != 0 {
		t.Errorf("Expected table to initially be empty, found %d rows\n", count)
	}

	expectedName := "ABC"
	expectedStep := 3

	err = log.Add(migrate.Migration{
		expectedName,
		expectedStep,
	})

	if err != nil {
		t.Fatal(err)
	}

	migrationQuery, err := db.Query("SELECT name, step FROM migrations;")

	if err != nil {
		t.Fatal(err)
	}
	defer migrationQuery.Close()

	var migrations []migrate.Migration
	var name string
	var step int

	for migrationQuery.Next() {

		err := migrationQuery.Scan(&name, &step)

		if err != nil {
			t.Fatal(err)
		}

		migrations = append(migrations, migrate.Migration{name, step})
	}

	if len(migrations) != 1 {
		t.Errorf("expected 1 migration, got %d\n", len(migrations))
	}

	if migrations[0].Name != expectedName {
		t.Errorf("expected migration name to be %s, got %s\n", expectedName, migrations[0].Name)
	}

	if migrations[0].Step != expectedStep {
		t.Errorf("expected migration step to be %d, got %d\n", expectedStep, migrations[0].Step)
	}

}

func TestSQLitePopReturnsMigrationAndRemovesFromTable(t *testing.T) {
	db, tearDown, err := sqliteDb()
	defer tearDown()

	if err != nil {
		t.Fatal(err)
	}

	log, err := migrate.NewLogSQLite(db)

	if err != nil {
		t.Fatal(err)
	}

	expectedName := "AAA1"
	expectedStep := 5

	_, err = db.Exec("INSERT INTO migrations (name, step) VALUES (?, ?);", expectedName, expectedStep)

	if err != nil {
		t.Fatal(err)
	}

	migration, err := log.Pop()

	if err != nil {
		t.Fatal(err)
	}

	if migration.Name != expectedName {
		t.Errorf("expected name %s, got %s\n", expectedName, migration.Name)
	}

	if migration.Step != expectedStep {
		t.Errorf("expected step %d, got %d\n", expectedStep, migration.Step)
	}

	countQuery := db.QueryRow("SELECT COUNT(*) FROM migrations;")

	var count int

	countQuery.Scan(&count)

	if count != 0 {
		t.Errorf("Expected table to be empty, found %d rows\n", count)
	}
}

// NextStep returns the next available step index
func TestSQLiteLastStepReturnsNextAvaiableIndex(t *testing.T) {
	db, tearDown, err := sqliteDb()
	defer tearDown()

	if err != nil {
		t.Fatal(err)
	}

	log, err := migrate.NewLogSQLite(db)

	if err != nil {
		t.Fatal(err)
	}

	expected := 5

	migrations := []migrate.Migration{
		{
			"aaa",
			4,
		},
		{
			"bbb",
			5,
		},
		{
			"ccc",
			5,
		},
	}

	for _, m := range migrations {
		_, err = db.Exec("INSERT INTO migrations (name, step) VALUES (?, ?);", m.Name, m.Step)

		if err != nil {
			t.Fatal(err)
		}
	}

	actual := log.LastStep()

	if expected != actual {
		t.Fatalf("Expected %d got %d", expected, actual)
	}
}
