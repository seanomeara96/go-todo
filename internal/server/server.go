package server

import (
	"go-todo/internal/db"
	"go-todo/internal/repositories"
	"go-todo/internal/server/cache"
	"go-todo/internal/server/logger"
	"go-todo/internal/server/renderer"
	"go-todo/internal/server/sessionstore"
	"go-todo/internal/services"
	"go-todo/web/delivery/http/handlers"
	"go-todo/web/delivery/http/router"
	"html/template"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Serve(env, port string) error {

	var logLevel logger.LogLevel = 0
	if os.Getenv("ENV") == "prod" {
		logLevel = 1
		logFile, err := logger.SetOutputToFile()
		if err != nil {
			return err
		}
		defer logFile.Close()
	}

	logger := logger.NewLogger(logLevel)

	db, err := db.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	store, err := sessionstore.GetSessionStore()
	if err != nil {
		return err
	}

	templateGlobPath := "./web/templates/**/*.html"
	tmpl := template.Must(template.ParseGlob(templateGlobPath))

	defaultExpiration := 5 * time.Minute
	cleanupInterval := 10 * time.Minute

	userCache := cache.NewUserCache(defaultExpiration, cleanupInterval, logger)
	todoCache := cache.NewTodoCache(defaultExpiration, cleanupInterval, logger)

	caches := &cache.Caches{
		UserCache: userCache,
		TodoCache: todoCache,
	}

	repository := repositories.NewRepository(db, logger)
	service := services.NewService(repository, caches, logger)
	renderer := renderer.NewRenderer(tmpl, logger)
	handler := handlers.NewHandler(service, store, renderer, logger)

	r := router.NewRouter(handler)

	logger.Info("Server started. Listening on http://localhost:" + port)
	return http.ListenAndServe(":"+port, r)
}
