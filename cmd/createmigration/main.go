package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jameswhoughton/migrate"
)

var dirFlag = flag.String("dir", "migrations", "set the directory in which to create migrations (default: migrations)")
var createPairFlag = flag.Bool("pair", false, "create a pair of migrations (up and down)")

func run(directory, name string, createPair bool) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err := os.Mkdir(directory, 0755)

		if err != nil {
			return err
		}
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	up, err := migrate.MakeMigration(directory, name, timestamp, "_up")

	if err != nil {
		return fmt.Errorf("up migration %s could not be created: %v", name, err)
	}

	migrations := []string{up}

	if createPair {
		down, err := migrate.MakeMigration(directory, name, timestamp, "_down")

		if err != nil {
			return fmt.Errorf("down migration %s could not be created: %v", name, err)
		}

		migrations = append(migrations, down)
	}

	for _, migration := range migrations {
		fmt.Printf("migration created: %s\n", migration)
	}

	return nil
}

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalln("createmigration expects one argument, the name of the migration")
	}

	name := flag.Args()[0]

	err := run(*dirFlag, name, *createPairFlag)

	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(0)
}
