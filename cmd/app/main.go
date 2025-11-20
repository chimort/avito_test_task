package main

import (
	"github.com/chimort/avito_test_task/iternal/app"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
)

func main() {
	log := logger.NewLogger("app", logger.LevelInfo)
	db := app.InitDB(log)
	app.RunMigrations(log, db)

	server := app.NewServer(log, db)
	server.Start(":8080")
}
