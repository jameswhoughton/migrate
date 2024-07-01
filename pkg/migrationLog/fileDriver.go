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

type FileDriver struct {
	FilePath   string
	migrations []Migration
}

func (m *Migration) string() string {
	return strconv.Itoa(m.Step) + "," + m.Name
}

func (ml *FileDriver) Load() error {
	logFile, err := os.OpenFile(ml.FilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModeAppend)

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

func (ml *FileDriver) Contains(search string) bool {
	for _, migration := range ml.migrations {
		if migration.Name == search {
			return true
		}
	}

	return false
}

func (ml *FileDriver) Count() int {
	return len(ml.migrations)
}

func (ml *FileDriver) Add(m Migration) error {
	file, err := os.OpenFile(ml.FilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return fmt.Errorf("cannot log file: %w", err)
	}

	defer file.Close()

	if _, err = file.WriteString(m.string() + "\n"); err != nil {
		return fmt.Errorf("cannot write to log file: %w", err)
	}

	ml.migrations = append(ml.migrations, m)

	return nil
}

func (ml *FileDriver) Pop() (Migration, error) {
	file, err := os.OpenFile(ml.FilePath, os.O_RDWR, 0644)
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
			fmt.Fprintln(file, migration.string())
		}
	}

	migration := ml.migrations[lastIndex]

	ml.migrations = ml.migrations[:lastIndex]

	return migration, nil
}

func (ml *FileDriver) LastStep() int {
	if len(ml.migrations) == 0 {
		return 0
	}

	lastIndex := len(ml.migrations) - 1

	return ml.migrations[lastIndex].Step
}

// Returns an instance of MigrationLog with migrations loaded
func (ml *FileDriver) Init() error {
	directory := filepath.Dir(ml.FilePath)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	if _, err := os.Stat(ml.FilePath); os.IsNotExist(err) {
		err := os.WriteFile(ml.FilePath, []byte{}, 0644)

		if err != nil {
			return fmt.Errorf("Error creating log file: %w", err)
		}
	}

	// Load the historic migrations from file
	err := ml.Load()

	if err != nil {
		return err
	}

	return nil
}
