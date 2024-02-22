package main

import (
	"go-todo/internal/db"
	"go-todo/internal/logger"
	"go-todo/internal/repositories"
	"go-todo/internal/server"
	"go-todo/internal/server/cache"
	"go-todo/internal/server/renderer"
	"go-todo/internal/server/sessionstore"
	"go-todo/internal/services"
	"go-todo/web/delivery/http/handlers"
	"go-todo/web/delivery/http/router"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load("./configs/.env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
		// You can choose to handle the error here or exit the program.
	}

	if os.Getenv("ENV") == "" || os.Getenv("PORT") == "" {
		log.Fatal("Expected a PORT and ENV var")
	}

	var logLevel logger.LogLevel = 0
	if os.Getenv("env") == "prod" {
		logLevel = 1
		logFile, err := logger.SetOutputToFile()
		if err != nil {
			panic(err)
		}
		defer logFile.Close()
	}

	logr := logger.NewLogger(logLevel)

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

	userCache := cache.NewUserCache(defaultExpiration, cleanupInterval)
	todoCache := cache.NewTodoCache(defaultExpiration, cleanupInterval)

	caches := &cache.Caches{UserCache: userCache, TodoCache: todoCache}

	repository := repositories.NewRepository(db)
	service := services.NewService(repository, caches)
	renderer := renderer.NewRenderer(tmpl)
	handler := handlers.NewHandler(service, store, renderer, logr)

	r := router.NewRouter(handler)
	if err = server.NewServer(r, logr).Serve(os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}
}
