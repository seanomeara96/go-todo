package server

import (
	"fmt"
	"go-todo/internal/db"
	"go-todo/internal/logger"
	"go-todo/internal/repositories"
	"go-todo/internal/server/cache"
	"go-todo/internal/server/renderer"
	"go-todo/internal/server/sessionstore"
	"go-todo/internal/services"
	"go-todo/web/delivery/http/handlers"
	"go-todo/web/delivery/http/router"
	"html/template"
	"net/http"
	"os"
	"time"

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

func DefaultServer(log *logger.Logger) *Server {

	db, err := db.Connect()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	useSecureSession := os.Getenv("env") == "prod"
	store, err := sessionstore.GetSessionStore(useSecureSession)
	if err != nil {
		panic(err)
	}

	templateGlobPath := "./web/templates/**/*.html"
	tmpl := template.Must(template.ParseGlob(templateGlobPath))

	defaultExpiration := 5 * time.Minute
	cleanupInterval := 10 * time.Minute

	userCache := cache.NewUserCache(defaultExpiration, cleanupInterval, log)
	todoCache := cache.NewTodoCache(defaultExpiration, cleanupInterval, log)

	caches := &cache.Caches{UserCache: userCache, TodoCache: todoCache}

	repository := repositories.NewRepository(db, log)
	service := services.NewService(repository, caches, log)
	renderer := renderer.NewRenderer(tmpl, log)
	handler := handlers.NewHandler(service, store, renderer, log)

	r := router.NewRouter(handler)
	return NewServer(r, log)
}
