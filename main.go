package migrate

import (
	"database/sql"
	"errors"
	"fmt"
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

type ErrorQuery struct {
	queryError error
	fileName   string
}

func (e ErrorQuery) Error() string {
	return "error executing query in " + e.fileName + ": " + e.queryError.Error()
}

type Migration struct {
	up   string
	down string
}

func (m *Migration) Name() string {
	nameRegex := regexp.MustCompile(`(.*)_up\.sql`)

	return nameRegex.FindString(m.up)
}

func Create(directory, name string) Migration {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	// Normalise names
	illegalCharacterRegexp := regexp.MustCompile(`[^a-zA-Z\d]+`)

	name = illegalCharacterRegexp.ReplaceAllString(name, "_")

	migrationName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), name)

	migration := Migration{
		up:   migrationName + "_up.sql",
		down: migrationName + "_down.sql",
	}

	os.WriteFile(directory+string(os.PathSeparator)+migration.up, []byte(""), 0644)
	os.WriteFile(directory+string(os.PathSeparator)+migration.down, []byte(""), 0644)

	return migration
}

func Migrate(driver *sql.DB, directory string, log *migrationLog.MigrationLog) error {
	migrations, err := os.ReadDir(directory)

	if err != nil {
		return err
	}

	newMigrations := false
	nameRegexp := regexp.MustCompile(`(.*)_up\.sql`)
	step := log.LastStep() + 1

	for _, migration := range migrations {
		fileName := migration.Name()

		if !isUpMigration(fileName, log) {
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
			return ErrorQuery{
				queryError: err,
				fileName:   fileName,
			}
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

func isUpMigration(fileName string, log *migrationLog.MigrationLog) bool {
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

func Rollback(driver *sql.DB, directory string, log *migrationLog.MigrationLog) error {
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
