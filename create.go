package migrate

import (
	"fmt"
	"os"
	"regexp"
)

func MakeMigration(directory, name, prefix, suffix string) (string, error) {
	// Normalise names
	illegalCharacterRegexp := regexp.MustCompile(`[^a-zA-Z\d]+`)

	name = illegalCharacterRegexp.ReplaceAllString(name, "_")

	migrationName := fmt.Sprintf("%s_%s_%s.sql", prefix, name, suffix)

	err := os.WriteFile(directory+string(os.PathSeparator)+migrationName, []byte(""), 0644)

	if err != nil {
		return "", err
	}

	return migrationName, nil
}
