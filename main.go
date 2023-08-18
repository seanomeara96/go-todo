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

	todoHandler := handlers.NewTodoHandler(todoService)
	pageHandler := handlers.NewPageHandler(tmpl, store)
	authHandler := handlers.NewAuthHandler(authService, store)

	r := mux.NewRouter()
	r.HandleFunc("/", pageHandler.Home).Methods("GET")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	r.HandleFunc("/todo/add", todoHandler.Add).Methods("POST")
	r.HandleFunc("/todo/remove", todoHandler.Remove).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":3000", r))
}
