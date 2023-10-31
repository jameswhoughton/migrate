package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	migrationLog, err := initMigrationLog("migrations/.log")

	if err != nil {
		log.Fatal(fmt.Errorf("Error initialising the log: %w\n", err))
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

type MigrationLog struct {
	filePath   string
	migrations []string
}

func (ml *MigrationLog) Name() string {
	return filepath.Base(ml.filePath)
}

func (ml *MigrationLog) Load() error {
	directory := filepath.Dir(ml.filePath)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	logFile, err := os.OpenFile(ml.filePath, os.O_APPEND|os.O_RDWR, os.ModeAppend)

	if err != nil {
		return errors.New("Cannot open log file: " + err.Error())
	}

	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)

	for scanner.Scan() {
		ml.migrations = append(ml.migrations, scanner.Text())
	}

	return nil
}

func (ml *MigrationLog) Contains(search string) bool {
	for _, migration := range ml.migrations {
		if migration == search {
			return true
		}
	}

	return false
}

func (ml *MigrationLog) Count() number {
	return len(ml.migrations)
}

func (ml *MigrationLog) Add(migration string) error {
	err := os.WriteFile(ml.filePath, []byte(migration), os.ModeAppend)

	if err != nil {
		return fmt.Errorf("Cannot write migration to log file: %w", err)
	}

	ml.migrations = append(ml.migrations, migration)

	return nil
}

func (ml *MigrationLog) Pop() (string, error) {
	file, err := os.OpenFile(ml.filePath, os.O_RDWR, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Empty the file
	file.Truncate(0)
	file.Seek(0, 0)

	lastIndex := len(ml.migrations) - 1

	for i, line := range ml.migrations {
		if i < lastIndex {
			fmt.Fprintln(file, line)
		}
	}

	migration := ml.migrations[lastIndex]

	ml.migrations = ml.migrations[:lastIndex-1]

	return migration, nil
}

func initMigrationLog(filePath string) (*MigrationLog, error) {
	log := MigrationLog{
		filePath: filePath,
	}

	// Load the historic migrations from file
	err := log.Load()

	if err != nil {
		return nil, err
	}

	return &log, nil
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

func migrate(driver *sql.DB, directory string, log *MigrationLog) error {
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

func rollback(driver *sql.DB, directory string, log *MigrationLog) error {
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
