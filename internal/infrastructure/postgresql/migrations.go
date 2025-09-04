package postgresql

import (
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations(pool *pgxpool.Pool) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return errFail("create migration driver: %w", err)
	}

	migrationsPath, err := findMigrationsPath()
	if err != nil {
		return errFail("find migrations: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		parseInput("file://%s", filepath.ToSlash(migrationsPath)),
		"postgres",
		driver,
	)
	if err != nil {
		return errFail("create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return errFail("run migrations: %w", err)
	}
	return nil
}

func findMigrationsPath() (string, error) {
	paths := []string{
		"/app/migrations",
		"./migrations",
		"../migrations",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errFail("migrations directory not found")
}
