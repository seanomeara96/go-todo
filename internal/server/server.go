package server

import (
	"fmt"
	"go-todo/internal/logger"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	router *mux.Router
	logger *logger.Logger
}

func NewServer(router *mux.Router, logger *logger.Logger) *Server {
	return &Server{router, logger}
}

func (s *Server) Serve(port string) error {
	s.logger.Info("Server started. Listening on http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, s.router); err != nil {
		return fmt.Errorf("Server failed to listen on %s. %w", port, err)
	}
	return nil
}
