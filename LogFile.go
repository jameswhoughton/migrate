package migrate

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type LogFile struct {
	FilePath   string
	Migrations []Migration
}

func (ml *LogFile) load() error {
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

		ml.Migrations = append(ml.Migrations, Migration{
			Name: parts[1],
			Step: Step,
		})
	}

	return nil
}

func (ml *LogFile) Contains(search string) bool {
	for _, migration := range ml.Migrations {
		if migration.Name == search {
			return true
		}
	}

	return false
}

func (ml *LogFile) Add(m Migration) error {
	file, err := os.OpenFile(ml.FilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return fmt.Errorf("cannot log file: %w", err)
	}

	defer file.Close()

	if _, err = file.WriteString(m.string() + "\n"); err != nil {
		return fmt.Errorf("cannot write to log file: %w", err)
	}

	ml.Migrations = append(ml.Migrations, m)

	return nil
}

func (ml *LogFile) Pop() (Migration, error) {
	file, err := os.OpenFile(ml.FilePath, os.O_RDWR, 0644)
	if err != nil {
		return Migration{}, err
	}
	defer file.Close()

	// Empty the file
	file.Truncate(0)
	file.Seek(0, 0)

	lastIndex := len(ml.Migrations) - 1

	for i, migration := range ml.Migrations {
		if i < lastIndex {
			fmt.Fprintln(file, migration.string())
		}
	}

	migration := ml.Migrations[lastIndex]

	ml.Migrations = ml.Migrations[:lastIndex]

	return migration, nil
}

func (ml *LogFile) LastStep() int {
	if len(ml.Migrations) == 0 {
		return 0
	}

	lastIndex := len(ml.Migrations) - 1

	return ml.Migrations[lastIndex].Step
}

// Returns an instance of MigrationLog with migrations loaded
func (ml *LogFile) Init() error {
	directory := filepath.Dir(ml.FilePath)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	if _, err := os.Stat(ml.FilePath); os.IsNotExist(err) {
		err := os.WriteFile(ml.FilePath, []byte{}, 0644)

		if err != nil {
			return fmt.Errorf("error creating log file: %w", err)
		}
	}

	// Load the historic migrations from file
	err := ml.load()

	if err != nil {
		return err
	}

	return nil
}

func NewLogFile(path string) (LogFile, error) {
	log := LogFile{
		FilePath: path,
	}

	err := log.Init()

	if err != nil {
		return LogFile{}, fmt.Errorf("failed to create file log: %w", err)
	}

	return log, nil
}
