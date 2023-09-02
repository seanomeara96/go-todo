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
	_ "github.com/mattn/go-sqlite3"
	"github.com/michaeljs1990/sqlitestore"
)

func main() {
	db, err := sql.Open("sqlite3", "main.db")
	if err != nil {
		log.Fatal(err)
	}

	store, err := sqlitestore.NewSqliteStore(
		"./main.db",
		"sessions",
		"/",
		3600,
		[]byte("<SecretKey>"),
	)
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

	tmpl, err := template.ParseGlob("./templates/**/*.html")
	if err != nil {
		log.Fatal(err)
	}

	todoRepo := repositories.NewTodoRepo(db)
	err = todoRepo.Init()
	if err != nil {
		log.Fatal(err)
	}
	userRepo := repositories.NewuserRepository(db)

	renderer := renderer.NewRenderer(tmpl)
	todoService := services.NewTodoService(todoRepo)
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService, store)
	todoHandler := handlers.NewTodoHandler(todoService, userService, renderer, store)
	pageHandler := handlers.NewPageHandler(userService, todoService, renderer, store)

	r := mux.NewRouter()
	r.HandleFunc("/", pageHandler.Home).Methods(http.MethodGet)
	r.HandleFunc("/signup", pageHandler.Signup).Methods(http.MethodGet)
	r.HandleFunc("/success", pageHandler.Success).Methods(http.MethodGet)
	r.HandleFunc("/cancel", pageHandler.Cancel).Methods(http.MethodGet)
	r.HandleFunc("/signup", userHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/upgrade", pageHandler.Upgrade).Methods(http.MethodGet)
	r.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/logout", authHandler.Logout).Methods(http.MethodGet)
	r.HandleFunc("/todo/add", todoHandler.Add).Methods(http.MethodPost)
	r.HandleFunc("/todo/update/status/{id}", todoHandler.UpdateStatus).Methods(http.MethodPost)
	r.HandleFunc("/todo/remove/{id}", todoHandler.Remove).Methods(http.MethodPost)
	r.HandleFunc("/create-checkout-session", userHandler.CreateCheckoutSession).Methods(http.MethodPost)
	log.Println("https://localhost:3000")
	//log.Fatal(http.ListenAndServe(":3000", r))
	log.Fatal(http.ListenAndServeTLS(":3000", "localhost.crt", "localhost.key", r))
}
