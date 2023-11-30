package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jameswhoughton/migrate/pkg/migrationLog"
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

	migrationLog, err := migrationLog.Init("migrations/.log")

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
	case args[0] == "rollback":
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

func migrate(driver *sql.DB, directory string, log *migrationLog.MigrationLog) error {
	migrations, err := os.ReadDir(directory)

	if err != nil {
		return err
	}

	newMigrations := false
	nameRegexp := regexp.MustCompile(`(.*)_up\.sql`)
	step := log.LastStep() + 1

	for _, migration := range migrations {
		fileName := migration.Name()

		if !isUpMigration(fileName, directory, log) {
			continue
		}

		migrationName := nameRegexp.FindStringSubmatch(fileName)

		// Ignore any migrations that have already run
		if log.Contains(migrationName[1]) {
			continue
		}

		// Run migration and add to log
		newMigrations = true

		query, err := os.ReadFile(directory + string(os.PathSeparator) + fileName)

		if err != nil {
			return err
		}

		_, err = driver.Exec(string(query))

		if err != nil {
			return err
		}

		err = log.Add(migrationName[1], step)

		if err != nil {
			return err
		}
	}

	if !newMigrations {
		return ErrorNoMigrations{}
	}

	return nil
}

func isUpMigration(fileName, directory string, log *migrationLog.MigrationLog) bool {
	// Ignore log file
	if fileName == log.Name() {
		return false
	}

	// Ignore down migrations
	if strings.HasSuffix(fileName, "_down.sql") {
		return false
	}

	return true
}

func rollback(driver *sql.DB, directory string, log *migrationLog.MigrationLog) error {
	if log.Count() == 0 {
		return errors.New("no migrations to roll back")
	}

	step := log.LastStep()

	for log.LastStep() == step {
		migration, err := log.Pop()

		if err != nil {
			return err
		}

		query, err := os.ReadFile(directory + string(os.PathSeparator) + migration.Name + "_down.sql")

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
