package main

import (
	"bufio"
	"database/sql"
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
		runMigrations(conn, "migrations")

	case args[0] == "create":
		if len(args) != 2 {
			log.Fatalln("Name required for new migration")
		}
	}
	createMigration("migrations", args[1])
}

func createMigration(directory, name string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	// Normalise names
	illegalCharacterRegexp := regexp.MustCompile(`[^a-zA-Z\d]+`)

	name = illegalCharacterRegexp.ReplaceAllString(name, "_")

	migrationName := fmt.Sprintf("%d_%s", time.Now().Nanosecond(), name)

	os.WriteFile(directory+string(os.PathSeparator)+migrationName+"_up.sql", []byte(""), 0644)
	os.WriteFile(directory+string(os.PathSeparator)+migrationName+"_down.sql", []byte(""), 0644)
}

func runMigrations(driver *sql.DB, directory string) error {
	logPath := directory + string(os.PathSeparator) + ".log"

	// Check for .log file and create if missing
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		os.WriteFile(logPath, []byte(""), 0644)
	}

	// Parse .log to work out last run migration
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR, os.ModeAppend)

	if err != nil {
		log.Fatalf("Cannot open log file: %s\n", err)
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

		_, err = logFile.Write([]byte(migration))

		if err != nil {
			return err
		}
	}
	//

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
