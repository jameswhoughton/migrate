package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jameswhoughton/migrate/pkg"
)

type ErrorNoMigrations struct{}

func (ErrorNoMigrations) Error() string {
	return "No new migrations to run"
}

func main() {
	// Accept database credentials

	// Handle arguments
	args := os.Args[1:]

	if len(args) > 2 {
		log.Fatalln("Too many arguments")
	}

	migrationLog, err := pkg.InitMigrationLog("migrations/.log")

	if err != nil {
		log.Fatal(fmt.Errorf("Error initialising the log: %w", err))
	}

	switch {
	case len(args) == 0:
		conn, _ := sql.Open("sqlite3", "test.db")
		migrate(conn, "migrations", migrationLog)

	case args[0] == "create":
		if len(args) != 2 {
			log.Fatalln("Name required for new migration")
		}
		create("migrations", args[1])
	case args[0] == "create":
		conn, _ := sql.Open("sqlite3", "test.db")
		rollback(conn, "migrations", migrationLog)
	}
}

type Migration struct {
	up   string
	down string
}

func (m *Migration) Name() string {
	nameRegex := regexp.MustCompile(`(.*)_up\.sql`)

	return nameRegex.FindString(m.up)
}

func create(directory, name string) Migration {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	// Normalise names
	illegalCharacterRegexp := regexp.MustCompile(`[^a-zA-Z\d]+`)

	name = illegalCharacterRegexp.ReplaceAllString(name, "_")

	migrationName := fmt.Sprintf("%d_%s", time.Now().Nanosecond(), name)

	migration := Migration{
		up:   migrationName + "_up.sql",
		down: migrationName + "_down.sql",
	}

	os.WriteFile(directory+string(os.PathSeparator)+migration.up, []byte(""), 0644)
	os.WriteFile(directory+string(os.PathSeparator)+migration.down, []byte(""), 0644)

	return migration
}

func migrate(driver *sql.DB, directory string, log *pkg.MigrationLog) error {
	migrations, err := os.ReadDir(directory)

	if err != nil {
		return err
	}

	var migrationsToRun []string

	for _, migration := range migrations {
		if migration.Name() == log.Name() {
			continue
		}

		if strings.HasSuffix(migration.Name(), "_down.sql") {
			continue
		}

		if log.Contains(migration.Name()) {
			continue
		}

		migrationsToRun = append(migrationsToRun, migration.Name())
	}

	if len(migrationsToRun) == 0 {
		return ErrorNoMigrations{}
	}

	// Loop over new migrations and execute writing to .log on success
	for _, migration := range migrationsToRun {
		query, err := os.ReadFile(directory + string(os.PathSeparator) + migration)

		if err != nil {
			return err
		}

		_, err = driver.Exec(string(query))

		if err != nil {
			return err
		}

		nameRegexp := regexp.MustCompile(`(.*)_up\.sql`)
		migrationName := nameRegexp.FindStringSubmatch(migration)

		err = log.Add(migrationName[1])

		if err != nil {
			return err
		}
	}

	return nil
}

func rollback(driver *sql.DB, directory string, log *pkg.MigrationLog) error {
	for log.Count() > 0 {
		migration, err := log.Pop()

		if err != nil {
			return err
		}

		query, err := os.ReadFile(directory + string(os.PathSeparator) + migration + "_down.sql")

		if err != nil {
			return err
		}

		_, err = driver.Exec(string(query))

		if err != nil {
			return err
		}
	}

	return nil
}
