package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
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
	case args[0] == "create":
		if len(args) != 2 {
			log.Fatalln("Name required for new migration")
		}

		createMigration("migrations", args[1])
	case len(args) == 0:
		//runMigrations(&osFS{}, conn, "migrations")
	}
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
	logFile, err := os.Open(logPath)

	if err != nil {
		log.Fatalf("Cannot open log file: %s\n", err)
	}

	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)
	var migrations []string

	for scanner.Scan() {
		migrations = append(migrations, scanner.Text())
	}

	if len(migrations) == 0 {
		return ErrorNoMigrations{}
	}

	// Loop over new migrations and execute writing to .log on success

	//

	return nil
}
