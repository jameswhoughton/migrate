package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type ErrorNoMigrations struct{}

func (ErrorNoMigrations) Error() string {
	return "No new migrations to run"
}

func main() {
	// Accept database credentials

	// Create migrations folder if it doesn't exist

	// Handle arguments
	args := os.Args[1:]

	if len(args) > 2 {
		log.Fatalln("Too many arguments")
	}

	switch {
	case len(args) == 0:
		conn, _ := sql.Open("sqlite3", "test.db")
		migrate(conn, "migrations")

	case args[0] == "create":
		if len(args) != 2 {
			log.Fatalln("Name required for new migration")
		}
		create("migrations", args[1])
	case args[0] == "create":
		conn, _ := sql.Open("sqlite3", "test.db")
		rollback(conn, "migrations")
	}
}

type Migration struct {
	up   string
	down string
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

func migrate(driver *sql.DB, directory string) error {
	logPath := directory + string(os.PathSeparator) + ".log"

	// Check for .log file and create if missing
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		os.WriteFile(logPath, []byte(""), 0644)
	}

	// Parse .log to work out last run migration
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR, os.ModeAppend)

	if err != nil {
		return errors.New("Cannot open log file: " + err.Error())
	}

	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)
	var runMigrations []string

	for scanner.Scan() {
		runMigrations = append(runMigrations, scanner.Text())
	}

	migrations, err := os.ReadDir(directory)

	if err != nil {
		return err
	}

	var migrationsToRun []string

	for _, migration := range migrations {
		if migration.Name() == ".log" {
			continue
		}

		if strings.HasSuffix(migration.Name(), "_down.sql") {
			continue
		}

		if sliceContains(runMigrations, migration.Name()) {
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

		nameRegexp := regexp.MustCompile(`[0-9]+_(.*)_up\.sql`)
		migrationName := nameRegexp.FindStringSubmatch(migration)

		_, err = logFile.Write([]byte(migrationName[1]))

		if err != nil {
			return err
		}
	}

	return nil
}

func rollback(driver *sql.DB, directory string) error {
	logPath := directory + string(os.PathSeparator) + ".log"

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return err
	}

	return nil
}

func sliceContains(s []string, e string) bool {
	for _, x := range s {
		if x == e {
			return true
		}
	}

	return false
}
