package migrationLog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type MigrationLog struct {
	filePath   string
	migrations []Migration
}

type Migration struct {
	Name string
	Step int
}

func (m *Migration) String() string {
	return strconv.Itoa(m.Step) + "," + m.Name
}

func (ml *MigrationLog) Name() string {
	return filepath.Base(ml.filePath)
}

func (ml *MigrationLog) Load() error {
	logFile, err := os.OpenFile(ml.filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModeAppend)

	if err != nil {
		return errors.New("Cannot open log file: " + err.Error())
	}

	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)

	// Parse the file to determine the total number of Steps
	for scanner.Scan() {
		fileLine := scanner.Text()
		parts := strings.Split(fileLine, ",")

		if len(parts) != 2 {
			return errors.New("log line malformed: " + fileLine)
		}

		Step, err := strconv.Atoi(parts[0])

		if err != nil {
			return errors.New("log line Step invalid: " + err.Error())
		}

		ml.migrations = append(ml.migrations, Migration{
			Name: parts[1],
			Step: Step,
		})
	}

	return nil
}

func (ml *MigrationLog) Contains(search string) bool {
	for _, migration := range ml.migrations {
		if migration.Name == search {
			return true
		}
	}

	return false
}

func (ml *MigrationLog) Count() int {
	return len(ml.migrations)
}

func (ml *MigrationLog) Add(Name string, Step int) error {
	m := Migration{
		Name: Name,
		Step: Step,
	}

	file, err := os.OpenFile(ml.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return fmt.Errorf("cannot log file: %w", err)
	}

	defer file.Close()

	if _, err = file.WriteString(m.String() + "\n"); err != nil {
		return fmt.Errorf("cannot write to log file: %w", err)
	}

	ml.migrations = append(ml.migrations, m)

	return nil
}

func (ml *MigrationLog) Pop() (Migration, error) {
	file, err := os.OpenFile(ml.filePath, os.O_RDWR, 0644)
	if err != nil {
		return Migration{}, err
	}
	defer file.Close()

	// Empty the file
	file.Truncate(0)
	file.Seek(0, 0)

	lastIndex := len(ml.migrations) - 1

	for i, migration := range ml.migrations {
		if i < lastIndex {
			fmt.Fprintln(file, migration.String())
		}
	}

	migration := ml.migrations[lastIndex]

	ml.migrations = ml.migrations[:lastIndex]

	return migration, nil
}

func (ml *MigrationLog) LastStep() int {
	if len(ml.migrations) == 0 {
		return 0
	}

	lastIndex := len(ml.migrations) - 1

	return ml.migrations[lastIndex].Step
}

// Returns an instance of MigrationLog with migrations loaded
func Init(filePath string) (*MigrationLog, error) {
	directory := filepath.Dir(filePath)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err := os.WriteFile(filePath, []byte{}, 0644)

		if err != nil {
			return nil, errors.New("Error creating log file: " + err.Error())
		}
	}

	log := MigrationLog{
		filePath: filePath,
	}

	// Load the historic migrations from file
	err := log.Load()

	if err != nil {
		return nil, err
	}

	return &log, nil
}
