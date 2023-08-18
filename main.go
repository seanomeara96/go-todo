package main

import (
	"database/sql"
	"encoding/gob"
	"go-todo/handlers"
	"go-todo/models"
	"go-todo/repositories"
	"go-todo/services"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/michaeljs1990/sqlitestore"
)

func main() {
	db, err := sql.Open("sqlite3", "main.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users(
		 id TEXT PRIMARY KEY UNIQUE NOT NULL,
		 name TEXT DEFAULT "",
		 email TEXT DEFAULT "",
		 password TEXT DEFAULT "")`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		description TEXT DEFAULT "",
		is_complete BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	store, err := sqlitestore.NewSqliteStore("./main.db", "sessions", "/", 3600, []byte("<SecretKey>"))
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

	tmpl, err := template.ParseGlob("./templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	todoRepo := repositories.NewTodoRepo(db)
	userRepo := repositories.NewuserRepository(db)

	todoService := services.NewTodoService(todoRepo)
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	todoHandler := handlers.NewTodoHandler(todoService, store)
	pageHandler := handlers.NewPageHandler(tmpl, store)
	authHandler := handlers.NewAuthHandler(authService, store)
	userHandler := handlers.NewUserHandler(userService)

	r := mux.NewRouter()
	r.HandleFunc("/", pageHandler.Home).Methods(http.MethodGet)
	r.HandleFunc("/signup", pageHandler.Signup).Methods(http.MethodGet)
	r.HandleFunc("/signup", userHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/logout", authHandler.Logout).Methods(http.MethodPost)
	r.HandleFunc("/todo/add", todoHandler.Add).Methods(http.MethodPost)
	r.HandleFunc("/todo/remove", todoHandler.Remove).Methods(http.MethodDelete)

	log.Fatal(http.ListenAndServe(":3000", r))
}
