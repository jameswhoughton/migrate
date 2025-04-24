package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

/*
Rollback reverses migrations that have been applied by executing the
corresponding rollback scripts in the correct order as determined by
the `log`.

Migrations are rolled back incrementally, for example, if 3 migrations
are applied and at a later date another 4 are applied, the first time
`Rollback` is called, only the most recent 4 migrations will be rolled back.
When called for a second time the next 3 migrations will be rolled back.

The name of the rollback file should match the following format:

`{numeric timestamp}_{name}_down.sql`

If a migration is missing a rollback file (e.g. a data change that is irreversible)
no action is taken and the next rollback in the group is processed.

If a rollback fails to run, an `ErrorQuery` error is returned.
*/
func Rollback(driver *sql.DB, directory fs.FS, log MigrationLog) error {
	step := log.LastStep()

	if step == 0 {
		return errors.New("no migrations to roll back")
	}

	for log.LastStep() == step {
		migration, err := log.Pop()

		if err != nil {
			return fmt.Errorf("Rollback: unable to pop migration from log: %v", err)
		}

		fileName := migration.Name + "_down.sql"

		if _, err := fs.Stat(directory, fileName); err != nil && errors.Is(err, os.ErrNotExist) {
			continue
		}

		query, err := fs.ReadFile(directory, fileName)

		if err != nil {
			return fmt.Errorf("Rollback: unable to read file: %v", err)
		}

		_, err = driver.Exec(string(query))

		if err != nil {
			return ErrorQuery{
				queryError: err,
				fileName:   fileName,
			}
		}
	}

	return nil
}
