package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

type Error string

func (e Error) Error() string { return string(e) }

var ERROR_NO_MIGRATIONS = Error("No new migrations to run")

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

		createMigration(&osFS{}, "migrations", args[1])
	case len(args) == 0:
		//runMigrations(&osFS{}, conn, "migrations")
	}
}

type fileSystem interface {
	Stat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool
	Mkdir(name string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
}

func createMigration(fs fileSystem, directory, name string) {
	if _, err := fs.Stat(directory); fs.IsNotExist(err) {
		fs.Mkdir(directory, 0755)
	}

	// Normalise names
	illegalCharacterRegexp := regexp.MustCompile(`[^a-zA-Z\d]+`)

	name = illegalCharacterRegexp.ReplaceAllString(name, "_")

	migrationName := fmt.Sprintf("%d_%s", time.Now().Nanosecond(), name)

	fs.WriteFile(directory+string(os.PathSeparator)+migrationName+"_up.sql", []byte(""), 0644)
	fs.WriteFile(directory+string(os.PathSeparator)+migrationName+"_down.sql", []byte(""), 0644)
}

type osFS struct{}

func (*osFS) Stat(name string) (os.FileInfo, error)     { return os.Stat(name) }
func (*osFS) IsNotExist(err error) bool                 { return os.IsNotExist(err) }
func (*osFS) Mkdir(name string, perm os.FileMode) error { return os.Mkdir(name, perm) }
func (*osFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func runMigrations(fs fileSystem, driver *sql.DB, directory string) error {
	// Check for .log file and create if missing
	if _, err := fs.Stat(directory + string(os.PathSeparator) + ".log"); fs.IsNotExist(err) {
		fs.WriteFile(directory+string(os.PathSeparator)+".log", []byte(""), 0644)
	}

	// Parse .log to work out last run migration

	var migrations []string

	if len(migrations) == 0 {
		return ERROR_NO_MIGRATIONS
	}

	// Loop over new migrations and execute writing to .log on success

	//

	return nil
}
