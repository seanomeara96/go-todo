package main

import (
	"database/sql"
	"fmt"
	"go-todo/cache"
	"go-todo/handlers"
	"go-todo/logger"
	"go-todo/renderer"
	"go-todo/repositories"
	"go-todo/services"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/michaeljs1990/sqlitestore"
	goCache "github.com/patrickmn/go-cache"
)

func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
		// You can choose to handle the error here or exit the program.
	}

	environment := os.Getenv("ENV")

	var logLevel logger.LogLevel = 0
	if environment == "prod" {
		logLevel = 1
	}

	logger := logger.NewLogger(logLevel)
	if environment == "prod" {
		fileName := "app.log"
		flag := os.O_APPEND | os.O_CREATE | os.O_WRONLY
		fileMode := fs.FileMode(0644)
		logFile, err := os.OpenFile(fileName, flag, fileMode)
		if err != nil {
			errMsg := fmt.Sprintf("Could not open file app.log %v", err)
			logger.Error(errMsg)
		}

		log.SetOutput(logFile)
		if err != nil {
			errMsg := fmt.Sprintf("Error opening log file: %v", err)
			logger.Error(errMsg)
		}
		defer logFile.Close()
	}

	driverName := "sqlite3"
	dataSourceName := "main.db"
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		logger.Error("Error connecting to db")

		debugMsg := fmt.Sprintf("sql.Open returned %v", err)
		logger.Debug(debugMsg)
		return
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		logger.Warning("Did not find a secret key in env vars")
	}

	endpoint := "./sessions.db"
	tableName := "sessions"
	path := "/"
	maxAge := 3600
	keyPairs := []byte(secretKey)
	store, err := sqlitestore.NewSqliteStore(endpoint, tableName, path, maxAge, keyPairs)
	if err != nil {
		logger.Error("Could not connect to session store")

		debugMsg := fmt.Sprintf("sqlitestore.NewSqliteStore returned %v", err)
		logger.Debug(debugMsg)
		return
	}

	sessionOptions := &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	if environment == "prod" {
		sessionOptions.Secure = true
	}

	store.Options = sessionOptions

	templateGlobPath := "./templates/**/*.html"
	tmpl, err := template.ParseGlob(templateGlobPath)
	if err != nil {
		logger.Error("Could not parse templates")

		debugMsg := fmt.Sprintf("template.ParseGlob returned %v", err)
		logger.Debug(debugMsg)

		return
	}

	c := goCache.New(5*time.Minute, 10*time.Minute)

	userCache := cache.NewUserCache(c, logger)
	todoCache := cache.NewTodoCache(c, logger)

	caches := &cache.Caches{
		UserCache: userCache,
		TodoCache: todoCache,
	}

	repository := repositories.NewRepository(db, logger)
	service := services.NewService(repository, caches, logger)
	renderer := renderer.NewRenderer(tmpl, logger)
	handler := handlers.NewHandler(service, store, renderer, logger)

	fs := http.FileServer(http.Dir("assets"))

	r := mux.NewRouter()
	r.Handle("/assets/", http.StripPrefix("/assets/", fs))
	r.HandleFunc("/", handler.HomePage).Methods(http.MethodGet)
	r.HandleFunc("/signup", handler.SignupPage).Methods(http.MethodGet)
	r.HandleFunc("/signup", handler.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/success", handler.SuccessPage).Methods(http.MethodGet)
	r.HandleFunc("/cancel", handler.CancelPage).Methods(http.MethodGet)
	r.HandleFunc("/upgrade", handler.UpgradePage).Methods(http.MethodGet)
	r.HandleFunc("/login", handler.Login).Methods(http.MethodPost)
	r.HandleFunc("/logout", handler.Logout).Methods(http.MethodGet)
	r.HandleFunc("/todo/add", handler.AddTodo).Methods(http.MethodPost)
	r.HandleFunc("/todo/update/description", handler.UpdateTodoDescription).Methods(http.MethodPost)
	r.HandleFunc("/todo/update/status/{id}", handler.UpdateTodoStatus).Methods(http.MethodPost)
	r.HandleFunc("/todo/remove/{id}", handler.RemoveTodo).Methods(http.MethodPost)
	r.HandleFunc("/create-checkout-session", handler.CreateCheckoutSession).Methods(http.MethodPost)
	r.HandleFunc("/manage-subscription", handler.CreateCustomerPortalSession).Methods(http.MethodGet)
	r.HandleFunc("/webhook", handler.HandleStripeWebhook).Methods(http.MethodPost)

	port := os.Getenv("PORT")
	if port == "" {
		logger.Error("Cant find port in env vars")
		return
	}

	logger.Info("Server started. Listening on http://localhost:" + port)
	logger.Error(http.ListenAndServe(":"+port, r).Error())
	//log.Fatal(http.ListenAndServeTLS(":3000", "localhost.crt", "localhost.key", r))
}
