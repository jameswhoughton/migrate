package migrationLog

import (
	"database/sql"
	"errors"
	"fmt"
)

type DBDriver struct {
	db sql.DB
}

func (d *DBDriver) Init() error {
	_, err := d.db.Exec("CREATE TABLE IF NOT EXISTS migrations (id INT NOT NULL auto_increment, name VARCHAR(100) NOT NULL, step INT NOT NULL)")

	if err != nil {
		return fmt.Errorf("could not create migrations table: %w", err)
	}

	return nil
}

func (d *DBDriver) Add(m Migration) error {
	_, err := d.db.Exec("INSERT INTO migrations (name, step) VALUES (?, ?)", m.Name, m.Step)

	if err != nil {
		return fmt.Errorf("unable to insert migration: %w", err)
	}

	return nil
}

func (d *DBDriver) Pop() (Migration, error) {
	row := d.db.QueryRow("SELECT FIRST id, name, step FROM migrations ORDER BY id DESC")

	var id int
	var name string
	var step int

	err := row.Scan(&id, &name, &step)

	if err != nil {
		return Migration{}, fmt.Errorf("unable to parse row: %w", err)
	}

	// Remove row
	d.db.Exec("DELETE FROM migrations WHERE id = ?", id)

	return Migration{
		Name: name,
		Step: step,
	}, nil
}

func (d *DBDriver) Contains(name string) bool {
	row := d.db.QueryRow("SELECT id FROM migrations WHERE name = ?", name)

	err := row.Scan()

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false
	}

	return true
}

func (d *DBDriver) LastStep() int {
	row := d.db.QueryRow("SELECT step FROM migrations ORDER BY id DESC")

	var step int

	err := row.Scan(&step)

	if err != nil {
		return 0
	}

	return step
}
