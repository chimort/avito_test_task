package app

import (
	"database/sql"
	"os"
	"time"

	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(log *logger.Logger) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dsn := "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"

	var db *sql.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Error("DB ping failed, retrying...", "error", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Error("Failed to connect to DB after retries", "error", err)
		os.Exit(1)
	}
	log.Info("DB connected")

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
