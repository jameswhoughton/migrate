# Migrate

Lightweight, DB agnostic migration tool which can be embedded in your application.

## Overview

### Migrations

Migrations can be stored anywhere although the default location is in a `migrations` directory at the root of your project. Each migration consists of two `.sql` files an up and a down, this is flexible, if you know you will never rollback a specific migration (e.g. irreversible data change) then the _down migration can be excluded. The migration files should follow the format `{unix timestamp in nanoseconds}_{migration name}_{up/down}.sql`. Migrations can be created manually or with the createmigration cli tool.

### Log

The migration log is used to keep track of which groups of migrations have been run. When `Migrate(...)` is called it will attempt to run all migrations (execute the `*_up.sql` files) which haven't been run in a single step. `Rollback(...)`, on the other hand, will roll back (execute the `*_down.sql` files) all migrations that have run in the previous step (not just the most recent migration).

### Log Drivers

At present the following migration log drivers are provided:
- File
- MySQL
- SQLite

For the file log driver, a file .log is created in the migrations directory this can be used if the DB you are using doesn't have a supported log driver.

For the DB log drivers, a new table `migrations` will be automatically created (if it doesn't already exist) when a new log instance is created.

All drivers implement the `MigrationLog` interface (`pkg/migrationLog/MigrationLog.go`).

## Usage

Install the dependency with `go get https://github.com/jameswhoughton/migrate`

### Using the File Log Driver
```go

import (
    "github.com/jameswhoughton/migrate"
    "github.com/jameswhoughton/migrate/pkg/migrationLog"
)

func main() {
    ...
    // Directory containing migrations
    migrationDir := "migrations"

    // Create an instance of the migration log
    log := migrationLog.NewLogFile(migrationDir + "/.log")

    // Create the connection to the DB
    db, _ := sql.Open("sqlite3", "test.db")

    // Call Migrate to run migrations
    migrate.Migrate(db, os.DirFS(migrationDir), log)
    ...
    // Call Rollback to reverse migrations
    migrate.Rollback(db, os.DirFS(migrationDir), log)
}
```

### Using the MySQL Log Driver
```go

import (
    "github.com/jameswhoughton/migrate"
    "github.com/jameswhoughton/migrate/pkg/migrationLog"
)

func main() {
    ...
    // Directory containing migrations
    migrationDir := "migrations"

    // Create the connection to the DB
    db, _ := sql.Open("mysql", "...")

    // Create an instance of the migration log
    log := migrationLog.NewLogMySQL(db)

    // Call Migrate to run migrations
    migrate.Migrate(db,os.DirFS( migrationDir), log)
    ...
    // Call Rollback to reverse migrations
    migrate.Rollback(db, os.DirFS(migrationDir), log)
}
```

### Using the SQLite Log Driver
```go

import (
    "github.com/jameswhoughton/migrate"
    "github.com/jameswhoughton/migrate/pkg/migrationLog"
)

func main() {
    ...
    // Directory containing migrations
    migrationDir := "migrations"

    // Create the connection to the DB
    db, _ := sql.Open("sqlite3", "...")

    // Create an instance of the migration log
    log := migrationLog.NewLogSQLite(db)

    // Call Migrate to run migrations
    migrate.Migrate(db, os.DirFS(migrationDir), log)
    ...
    // Call Rollback to reverse migrations
    migrate.Rollback(db, os.DirFS(migrationDir), log)
}
```