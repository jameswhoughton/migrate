package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type MigrationLog struct {
	filePath   string
	migrations []string
}

func (ml *MigrationLog) Name() string {
	return filepath.Base(ml.filePath)
}

func (ml *MigrationLog) Load() error {
	directory := filepath.Dir(ml.filePath)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	logFile, err := os.OpenFile(ml.filePath, os.O_APPEND|os.O_RDWR, os.ModeAppend)

	if err != nil {
		return errors.New("Cannot open log file: " + err.Error())
	}

	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)

	for scanner.Scan() {
		ml.migrations = append(ml.migrations, scanner.Text())
	}

	return nil
}

func (ml *MigrationLog) Contains(search string) bool {
	for _, migration := range ml.migrations {
		if migration == search {
			return true
		}
	}

	return false
}

func (ml *MigrationLog) Count() int {
	return len(ml.migrations)
}

func (ml *MigrationLog) Add(migration string) error {
	err := os.WriteFile(ml.filePath, []byte(migration), os.ModeAppend)

	if err != nil {
		return fmt.Errorf("Cannot write migration to log file: %w", err)
	}

	ml.migrations = append(ml.migrations, migration)

	return nil
}

func (ml *MigrationLog) Pop() (string, error) {
	file, err := os.OpenFile(ml.filePath, os.O_RDWR, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Empty the file
	file.Truncate(0)
	file.Seek(0, 0)

	lastIndex := len(ml.migrations) - 1

	for i, line := range ml.migrations {
		if i < lastIndex {
			fmt.Fprintln(file, line)
		}
	}

	migration := ml.migrations[lastIndex]

	ml.migrations = ml.migrations[:lastIndex-1]

	return migration, nil
}

func initMigrationLog(filePath string) (*MigrationLog, error) {
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
