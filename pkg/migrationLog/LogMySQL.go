package migrationLog

import (
	"database/sql"
	"errors"
	"fmt"
)

type LogMySQL struct {
	db *sql.DB
}

func (d *LogMySQL) init() error {
	_, err := d.db.Exec("CREATE TABLE IF NOT EXISTS migrations (id INT PRIMARY KEY auto_increment, name VARCHAR(100) NOT NULL, step INT NOT NULL);")

	if err != nil {
		return fmt.Errorf("could not create migrations table: %w", err)
	}

	return nil
}

func (d *LogMySQL) Add(m Migration) error {
	_, err := d.db.Exec("INSERT INTO migrations (name, step) VALUES (?, ?)", m.Name, m.Step)

	if err != nil {
		return fmt.Errorf("unable to insert migration: %w", err)
	}

	return nil
}

func (d *LogMySQL) Pop() (Migration, error) {
	row := d.db.QueryRow("SELECT id, name, step FROM migrations ORDER BY id DESC LIMIT 1")

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

func (d *LogMySQL) Contains(name string) bool {
	row := d.db.QueryRow("SELECT id FROM migrations WHERE name = ?", name)

	err := row.Scan()

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false
	}

	return true
}

func (d *LogMySQL) LastStep() int {
	row := d.db.QueryRow("SELECT step FROM migrations ORDER BY id DESC")

	var step int

	err := row.Scan(&step)

	if err != nil {
		return 0
	}

	return step
}

func NewLogMySQL(db *sql.DB) (LogMySQL, error) {
	log := LogMySQL{
		db: db,
	}

	err := log.init()

	if err != nil {
		return LogMySQL{}, fmt.Errorf("failed to create MySQL log: %w", err)
	}

	return log, nil
}
