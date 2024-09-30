package migrate

import "strconv"

type Migration struct {
	Name string
	Step int
}

func (m *Migration) string() string {
	return strconv.Itoa(m.Step) + "," + m.Name
}

type MigrationLog interface {
	Init() error
	Add(m Migration) error
	Pop() (Migration, error)
	Contains(name string) bool
	LastStep() int
}
