package app

import (
	"database/sql"
	"os"

	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(log *logger.Logger, db *sql.DB) {

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Error("failed to create migration driver", "error", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Error("failed to create migrate instance", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("No migrations to apply")
		} else {
			log.Error("failed to apply migrations", "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("Migrations applied")
	}
}
