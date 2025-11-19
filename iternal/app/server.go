package app

import (
	"os"

	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Server struct {
	echo *echo.Echo
	log  *logger.Logger
}

func NewServer(log *logger.Logger) *Server {
	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(200, "pong")
	})

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
