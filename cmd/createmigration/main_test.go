package main

import (
	"os"
	"testing"
)

const MIGRATION_DIR = "migrations_test"

// creates migrations dir if doesn't exist
func TestCreatesMigrationDirectoryIfMissing(t *testing.T) {
	defer os.RemoveAll(MIGRATION_DIR)

	run(MIGRATION_DIR, "test", false)

	if _, err := os.Stat(MIGRATION_DIR); os.IsNotExist(err) {
		t.Fatal("migrations directory not found")
	}
}
