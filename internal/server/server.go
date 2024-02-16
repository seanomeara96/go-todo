package server

import (
	"go-todo/internal/db"
	"go-todo/internal/server/cache"
	"go-todo/internal/server/logger"
	"go-todo/internal/server/renderer"
	"go-todo/internal/server/repositories"
	"go-todo/internal/server/services"
	"go-todo/internal/server/sessionstore"
	"go-todo/web/delivery/http/handlers"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func Serve() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
		// You can choose to handle the error here or exit the program.
		return
	}

	if os.Getenv("ENV") == "" || os.Getenv("PORT") == "" {
		log.Fatal("Expected a PORT and ENV var")
		return
	}

	var logLevel logger.LogLevel = 0
	if os.Getenv("ENV") == "prod" {
		logLevel = 1
		logFile, err := logger.SetOutputToFile()
		if err != nil {
			log.Fatal("Could not set output file")
			return
		}
		defer logFile.Close()
	}

	logger := logger.NewLogger(logLevel)

	db, err := db.Connect()
	if err != nil {
		logger.Error("Error connecting to db")
		logger.Debug(err.Error())
		return
	}
	defer db.Close()

	store, err := sessionstore.GetSessionStore()
	if err != nil {
		logger.Error("Could not connect to session store")
		logger.Debug(err.Error())
		return
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

	r := routes.Router(handler)

	port := os.Getenv("PORT")
	if port == "" {
		logger.Error("Cant find port in env vars")
		return
	}

	logger.Info("Server started. Listening on http://localhost:" + port)
	logger.Error(http.ListenAndServe(":"+port, r).Error())
}
