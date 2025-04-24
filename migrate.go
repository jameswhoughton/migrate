/*
Lightweight, DB agnostic migration tool.
*/
package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
)

type ErrorQuery struct {
	queryError error
	fileName   string
}

func (e ErrorQuery) Error() string {
	return "error executing query in " + e.fileName + ": " + e.queryError.Error()
}

/*
Migrate executes all migrations that haven't previously run.

Each time Migrate runs the `Step` is incremented, all migrations that are
successfully executed are added to the log with the same step (this allows them to
be rolled back in the same group (see Rollback for more detail).

The name of the migration file should match the following format:

`{numeric timestamp}_{name}{_up}?.sql`

The `_up` suffix is optional as in some cases rollback scripts may not be required,
even if there is a rollback script `_up` is not required (but maybe useful for clarity).

Migrations are executed in ascending order.

If a rollback fails to run, an `ErrorQuery` error is returned.
*/
func Migrate(driver *sql.DB, directory fs.FS, log MigrationLog) error {
	migrations, err := fs.Glob(directory, `*.sql`)

	if err != nil {
		return fmt.Errorf("Migrate: unable to retrieve migration files: %v", err)
	}

	// ensure migrations are ordered
	sort.Strings(migrations)

	nameRegexp := regexp.MustCompile(`(.*?)(_up|_down)?\.sql`)
	step := log.LastStep() + 1

	for _, migration := range migrations {
		nameParts := nameRegexp.FindStringSubmatch(migration)

		// Ignore any down migrations
		if len(nameParts) == 3 && nameParts[2] == "_down" {
			continue
		}

		// Ignore any migrations that have already run
		if log.Contains(nameParts[1]) {
			continue
		}

		query, err := fs.ReadFile(directory, migration)

		if err != nil {
			return fmt.Errorf("Migrate: unable to read migration '%s': %v", migration, err)
		}

		_, err = driver.Exec(string(query))

		if err != nil {
			return ErrorQuery{
				queryError: err,
				fileName:   migration,
			}
		}

		err = log.Add(Migration{
			Name: nameParts[1],
			Step: step,
		})

		if err != nil {
			return fmt.Errorf("Migrate: unable to add migration '%s' to log: %v", migration, err)
		}
	}

	return nil
}
