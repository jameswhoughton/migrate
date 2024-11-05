package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"regexp"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

type ErrorQuery struct {
	queryError error
	fileName   string
}

func (e ErrorQuery) Error() string {
	return "error executing query in " + e.fileName + ": " + e.queryError.Error()
}

func Migrate(driver *sql.DB, directory fs.FS, log MigrationLog) error {
	migrations, err := fs.Glob(directory, `*.sql`)

	if err != nil {
		return err
	}

	// ensure migrations are ordered
	sort.Strings(migrations)

	nameRegexp := regexp.MustCompile(`(.*?)(_up|_down)?\.sql`)
	step := log.LastStep() + 1

	for _, migration := range migrations {
		nameParts := nameRegexp.FindStringSubmatch(migration)

		// Ignore any down migrations
		fmt.Print(nameParts)
		if len(nameParts) == 3 && nameParts[2] == "_down" {
			continue
		}

		// Ignore any migrations that have already run
		if log.Contains(nameParts[1]) {
			continue
		}

		query, err := fs.ReadFile(directory, migration)

		if err != nil {
			return err
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
			return err
		}
	}

	return nil
}
