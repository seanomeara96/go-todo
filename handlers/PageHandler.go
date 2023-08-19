package handlers

import (
	"go-todo/models"
	"go-todo/services"
	"html/template"
	"net/http"

	"github.com/michaeljs1990/sqlitestore"
)

type PageHandler struct {
	tmpl        *template.Template
	store       *sqlitestore.SqliteStore
	userService *services.UserService
	todoService *services.TodoService
}

func NewPageHandler(userService *services.UserService, todoService *services.TodoService, tmpl *template.Template, store *sqlitestore.SqliteStore) *PageHandler {
	return &PageHandler{
		tmpl:        tmpl,
		store:       store,
		userService: userService,
		todoService: todoService,
	}
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)

	var list []*models.Todo
	if user != nil {
		list, err = h.todoService.GetUserTodoList(user.ID)
		if err != nil {
			http.Error(w, "could not get users list of todos", http.StatusInternalServerError)
			return
		}
	}

	type HomePageData struct {
		User  *models.User
		Todos []*models.Todo
	}

	var homePageData HomePageData = HomePageData{
		User:  user,
		Todos: list,
	}
	err = h.tmpl.ExecuteTemplate(w, "home", homePageData)
	if err != nil {
		http.Error(w, "Could not render homepage", http.StatusInternalServerError)
		return
	}
}

func (h *PageHandler) Signup(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	err = h.tmpl.ExecuteTemplate(w, "signup", nil)
	if err != nil {
		http.Error(w, "could not render sigup page", http.StatusInternalServerError)
		return
	}

}
