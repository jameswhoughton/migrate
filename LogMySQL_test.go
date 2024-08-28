package migrate

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func mysqlDb() (*sql.DB, func(), error) {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:8022)/testing")

	if err != nil {
		return nil, nil, err
	}

	return db, func() {
		db.Exec("DROP TABLE migrations")
	}, nil

}

func TestNewLogMySQLCreatesMigrationsTable(t *testing.T) {
	db, tearDown, err := mysqlDb()
	defer tearDown()

	if err != nil {
		t.Fatal(err)
	}

	_, err = NewLogMySQL(db)

	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'testing' AND table_name = 'migrations';")

	var count int

	row.Scan(&count)

	if count != 1 {
		t.Errorf("Migration table missing")
	}
}

// Contains() returns true if the given migration exists in the log
func TestMySQLContainsReturnsTheCorrectResult(t *testing.T) {
	db, tearDown, err := mysqlDb()

	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		name       string
		migrations []Migration
		search     string
		expected   bool
	}

	cases := []testCase{
		{
			name: "search in list",
			migrations: []Migration{
				{"a", 0},
				{"b", 0},
				{"c", 0},
			},
			search:   "a",
			expected: true,
		},
		{
			name:       "empty migrations",
			migrations: []Migration{},
			search:     "a",
			expected:   false,
		},
		{
			name: "partial search",
			migrations: []Migration{
				{"migration A", 0},
				{"migration B", 0},
			},
			search:   "migration",
			expected: false,
		},
		{
			name: "different steps",
			migrations: []Migration{
				{"migration A", 0},
				{"migration B", 1},
			},
			search:   "migration B",
			expected: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			defer tearDown()

			migrationLog, err := NewLogMySQL(db)

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

func TestMySQLAddInsertsMigrationIntoTable(t *testing.T) {
	db, tearDown, err := mysqlDb()
	defer tearDown()

	if err != nil {
		t.Fatal(err)
	}

	log, err := NewLogMySQL(db)

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

	err = log.Add(Migration{
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

	var migrations []Migration
	var name string
	var step int

	for migrationQuery.Next() {

		err := migrationQuery.Scan(&name, &step)

		if err != nil {
			t.Fatal(err)
		}

		migrations = append(migrations, Migration{name, step})
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

func TestMySQLPopReturnsMigrationAndRemovesFromTable(t *testing.T) {
	db, tearDown, err := mysqlDb()
	defer tearDown()

	if err != nil {
		t.Fatal(err)
	}

	log, err := NewLogMySQL(db)

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
func TestMySQLLastStepReturnsNextAvaiableIndex(t *testing.T) {
	db, tearDown, err := mysqlDb()
	defer tearDown()

	if err != nil {
		t.Fatal(err)
	}

	log, err := NewLogMySQL(db)

	if err != nil {
		t.Fatal(err)
	}

	expected := 5

	migrations := []Migration{
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
