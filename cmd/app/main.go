package main

import (
	"github.com/chimort/avito_test_task/iternal/app"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
)

func main() {
	log := logger.NewLogger("app", logger.LevelInfo)
	app.RunMigrations(log)

	server := app.NewServer(log)
	server.Start(":8080")
}
