package migrate

import "strconv"

/*
Representation of a migration, the name should match the name of the
file without the suffix, for example, a migration with the file name
`123_create_table_up.sql` will have the Migration.Name `123_create_table`

The step is a numeric representation of the group, migrations are grouped
based upon when they are executed.
*/
type Migration struct {
	Name string
	Step int
}

func (m *Migration) string() string {
	return strconv.Itoa(m.Step) + "," + m.Name
}

/*
Interface representing the log, responsible for keeping track of which migrations
have run and in what order/grouping.

In most cases the log should be stored in the database, the package includes
implementations for the following DBMS:

  - SQLite
  - MySQL

Alternatively there is also a File implementation (LogFile) or you are free
to create your own type for whichever DBMS you need.
*/
type MigrationLog interface {
	Init() error
	Add(m Migration) error
	Pop() (Migration, error)
	Contains(name string) bool
	LastStep() int
}
