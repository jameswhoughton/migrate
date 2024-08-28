package migrate

import (
	"database/sql"
	"errors"
	"io/fs"
)

func Rollback(driver *sql.DB, directory fs.FS, log MigrationLog) error {
	step := log.LastStep()

	if step == 0 {
		return errors.New("no migrations to roll back")
	}

	for log.LastStep() == step {
		migration, err := log.Pop()

		if err != nil {
			return err
		}

		query, err := fs.ReadFile(directory, migration.Name+"_down.sql")

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
