/*
Simple CLI tool to create migration and (optionally) its corresponding rollback.

This is the suggested method to create script files although there is nothing
to prevent you creating them manually.

The command accepts a single argument which is the name of the script. Any non-numeric
characters will be replaced with an underscore, for example, an argument of `create users table`
with the --pair option will create a migration with the name 444_create_users_table_up.sql and
a rollback with the name 444_create_users_table_down.sql (where 444 is the current timestamp).

The optional `--dir` option specifies the directory to create the scripts (relative
to the run path), the default value is 'migrations'.

The optional `--pair` option will create both a migration and a rollback script.
*/
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
var helpFlag = flag.Bool("help", false, "help")

func run(directory, name string, createPair bool) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err := os.Mkdir(directory, 0755)

		if err != nil {
			return err
		}
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	suffix := ""

	if createPair {
		suffix = "up"
	}

	up, err := migrate.MakeMigration(directory, name, timestamp, suffix)

	if err != nil {
		return fmt.Errorf("up migration %s could not be created: %v", name, err)
	}

	migrations := []string{up}

	if createPair {
		down, err := migrate.MakeMigration(directory, name, timestamp, "down")

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

func showHelp() {
	fmt.Print(`Create Migration CLI Tool

Description
  Tool to create migration and rollback script files which are 
  compatible with https://github.com/jameswhoughton/migrate

Usage:
  createmigration [--pair] [--dir=] name

Flags:
  --pair	Create both a migration and a rollback script, 
		if omitted, only the migration will be created.
  --dir		Specify the directory in which to save the scripts,
		the path should be relative to the command location.
		The default value is 'migrations'.
`)
}

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()

		os.Exit(0)
	}

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
