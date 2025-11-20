package app

import (
	"database/sql"
	"os"
	"time"

	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	_ "github.com/lib/pq"
)

func InitDB(log *logger.Logger) *sql.DB {
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
				log.Info("DB connected")
				return db
			}
		}
		log.Error("DB ping failed, retrying...", "error", err)
		time.Sleep(2 * time.Second)
	}

	log.Error("Failed to connect to DB after retries", "error", err)
	os.Exit(1)
	return nil
}
