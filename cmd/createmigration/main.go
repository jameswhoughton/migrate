package main

import (
	"flag"
	"log"

	"github.com/jameswhoughton/migrate"
)

var dirFlag = flag.String("dir", "migrations", "set the directory in which to create migrations (default: migrations)")

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalln("createmigration expects one argument, the name of the migration")
	}

	name := flag.Args()[0]

	migrate.Create(*dirFlag, name)
}
