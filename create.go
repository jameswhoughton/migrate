package migrate

import (
	"fmt"
	"os"
	"regexp"
)

/*
MakeMigration creates a new, empty script file compatible with Migrate(...)
and Rollback(...).

Note this function is used internally in the createmigration CLI, that is the
suggested way to create migrations, it has been made publically available for
advanced use only.

Any non-alphanumeric characters are replaced with an underscore.

The prefix is used to determine the order of execution, in most cases a timestamp
is a good choice, whatever is chosen, the migration and corresponding rollback scripts
should have the same prefix.

The suffix can be blank or 'up' (if creating migrations) but when creating rollbacks
the suffix should always be 'down'.
*/
func MakeMigration(directory, name, prefix, suffix string) (string, error) {
	// Normalise names
	illegalCharacterRegexp := regexp.MustCompile(`[^a-zA-Z\d]+`)

	name = illegalCharacterRegexp.ReplaceAllString(name, "_")

	migrationName := ""

	if prefix != "" {
		migrationName += prefix + "_"
	}

	migrationName += name

	if suffix != "" {
		migrationName += "_" + suffix
	}

	migrationName += ".sql"

	err := os.WriteFile(directory+string(os.PathSeparator)+migrationName, []byte(""), 0644)

	if err != nil {
		return "", fmt.Errorf("MakeMigration: unable to write file %s to directory %s: %v", migrationName, directory, err)
	}

	return migrationName, nil
}
