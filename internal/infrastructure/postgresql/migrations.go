package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	projectRoot, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("could not get project root: %w", err)
	}

	migrationsPath := fmt.Sprintf("file://%s/migrations", projectRoot)

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)

	if err != nil {
		return fmt.Errorf("could not create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

func getProjectRoot() (string, error) {
	containerPaths := []string{
		"/app",
		"/migrations",
	}

	for _, path := range containerPaths {
		migrationsDir := path
		if path != "/migrations" {
			migrationsDir = filepath.Join(path, "migrations")
		}

		if _, err := os.Stat(migrationsDir); err == nil {
			if path == "/migrations" {
				return "/", nil
			}
			return path, nil
		}
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	possibleLocalPaths := []string{
		currentDir,
		filepath.Join(currentDir, "..", ".."),
	}

	for _, path := range possibleLocalPaths {
		migrationsDir := filepath.Join(path, "migrations")
		if _, err := os.Stat(migrationsDir); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found")
}
