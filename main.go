package main

import (
	"database/sql"
	"encoding/gob"
	"go-todo/handlers"
	"go-todo/models"
	"go-todo/renderer"
	"go-todo/repositories"
	"go-todo/services"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/michaeljs1990/sqlitestore"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
		// You can choose to handle the error here or exit the program.
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

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	//By calling gob.Register(&CustomData{}), you're letting the gob package know how to encode and decode instances of your CustomData struct.
	gob.Register(models.User{})

	templateGlobPath := "./templates/**/*.html"
	tmpl, err := template.ParseGlob(templateGlobPath)
	if err != nil {
		log.Fatal(err)
	}

	repository := repositories.NewRepository(db)
	service := services.NewService(repository)
	renderer := renderer.NewRenderer(tmpl)
	handler := handlers.NewHandler(service, store, renderer)

	r := mux.NewRouter()
	r.HandleFunc("/", handler.Home).Methods(http.MethodGet)
	r.HandleFunc("/signup", handler.Signup).Methods(http.MethodGet)
	r.HandleFunc("/success", handler.Success).Methods(http.MethodGet)
	r.HandleFunc("/cancel", handler.Cancel).Methods(http.MethodGet)
	r.HandleFunc("/signup", handler.Create).Methods(http.MethodPost)
	r.HandleFunc("/upgrade", handler.Upgrade).Methods(http.MethodGet)
	r.HandleFunc("/login", handler.Login).Methods(http.MethodPost)
	r.HandleFunc("/logout", handler.Logout).Methods(http.MethodGet)
	r.HandleFunc("/todo/add", handler.Add).Methods(http.MethodPost)
	r.HandleFunc("/todo/update/status/{id}", handler.UpdateStatus).Methods(http.MethodPost)
	r.HandleFunc("/todo/remove/{id}", handler.Remove).Methods(http.MethodPost)
	r.HandleFunc("/create-checkout-session", handler.CreateCheckoutSession).Methods(http.MethodPost)
	r.HandleFunc("/webhook", handler.HandleStripeWebhook).Methods(http.MethodPost)

	log.Println("http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", r))
	//log.Fatal(http.ListenAndServeTLS(":3000", "localhost.crt", "localhost.key", r))
}
