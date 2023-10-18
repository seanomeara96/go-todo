package main

import (
	"go-todo/internal/cache"
	"go-todo/internal/db"
	"go-todo/internal/handlers"
	"go-todo/internal/logger"
	"go-todo/internal/renderer"
	"go-todo/internal/repositories"
	"go-todo/internal/services"
	"go-todo/web/routes"
	"go-todo/web/sessionstore"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
		// You can choose to handle the error here or exit the program.
	}

	if os.Getenv("ENV") == "" || os.Getenv("PORT") == "" {
		log.Fatal("Expected a PORT and ENV var")
	}

	var logLevel logger.LogLevel = 0
	if os.Getenv("ENV") == "prod" {
		logLevel = 1
		logFile, _ := logger.SetOutputToFile()
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
