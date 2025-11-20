package app

import (
	"database/sql"
	"os"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/handlers"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/chimort/avito_test_task/iternal/service"
	"github.com/labstack/echo/v4"
)

type Server struct {
	echo *echo.Echo
	log  *logger.Logger
}

func NewServer(log *logger.Logger, db *sql.DB) *Server {
	e := echo.New()

	repo := repository.NewUserRepository(db)
	userService := service.NewUserService(repo, log)
	h := handlers.NewHandlers(userService, log)

	api.RegisterHandlers(e, h)

	return &Server{
		echo: e,
		log:  log,
	}
}

func (s *Server) Start(port string) {
	s.log.Info("Server started on " + port)
	if err := s.echo.Start(port); err != nil {
		s.log.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
