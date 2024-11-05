package migrate

import (
	"os"
	"regexp"
)

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
		return "", err
	}

	return migrationName, nil
}
