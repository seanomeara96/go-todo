package main

import (
	"database/sql"
	"go-todo/cache"
	"go-todo/handlers"
	"go-todo/logger"
	"go-todo/renderer"
	"go-todo/repositories"
	"go-todo/services"
	"html/template"
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

	if os.Getenv("env") == "prod" {
		logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logFile)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		defer logFile.Close()

	}

	db, err := sql.Open("sqlite3", "main.db")
	if err != nil {
		log.Fatal(err)
	}

	secretKey := []byte(os.Getenv("SECRET_KEY"))
	store, err := sqlitestore.NewSqliteStore("./sessions.db", "sessions", "/", 3600, secretKey)
	if err != nil {
		log.Fatal(err)
	}

	sessionOptions := &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	if os.Getenv("env") == "prod" {
		sessionOptions.Secure = true
	}

	store.Options = sessionOptions

	templateGlobPath := "./templates/**/*.html"
	tmpl, err := template.ParseGlob(templateGlobPath)
	if err != nil {
		log.Fatal(err)
	}

	c := goCache.New(5*time.Minute, 10*time.Minute)

	var logLevel logger.LogLevel = 0
	if os.Getenv("env") == "prod" {
		logLevel = 1
	}
	logger := logger.NewLogger(logLevel)

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

	logger.Info("Server started. Listening on http://localhost:3000")
	logger.Error(http.ListenAndServe(":3000", r).Error())
	//log.Fatal(http.ListenAndServeTLS(":3000", "localhost.crt", "localhost.key", r))
}
