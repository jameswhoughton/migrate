package migrationLog

type Migration struct {
	Name string
	Step int
}

type MigrationLog interface {
	Init() error
	Add(m Migration) error
	Pop() (Migration, error)
	Contains(name string) bool
	LastStep() int
}
