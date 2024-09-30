package migrate_test

import "github.com/jameswhoughton/migrate"

type testLog struct {
	store []migrate.Migration
}

func (ml *testLog) Init() error {
	return nil
}
func (ml *testLog) Add(m migrate.Migration) error {
	ml.store = append(ml.store, m)

	return nil
}
func (ml *testLog) Pop() (migrate.Migration, error) {
	lastIndex := len(ml.store) - 1
	migration := ml.store[lastIndex]

	ml.store = ml.store[:lastIndex]

	return migration, nil
}
func (ml *testLog) Contains(name string) bool {
	for _, migration := range ml.store {
		if migration.Name == name {
			return true
		}
	}

	return false
}
func (ml *testLog) LastStep() int {
	if len(ml.store) == 0 {
		return 0
	}

	lastIndex := len(ml.store) - 1

	return ml.store[lastIndex].Step
}

func newTestLog() testLog {
	return testLog{}
}
