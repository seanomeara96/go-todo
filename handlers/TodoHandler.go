package handlers

import (
	"go-todo/services"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/michaeljs1990/sqlitestore"
)

type TodoHandler struct {
	todoService *services.TodoService
	store       *sqlitestore.SqliteStore
	tmpl        *template.Template
}

func NewTodoHandler(todoService *services.TodoService, tmpl *template.Template, store *sqlitestore.SqliteStore) *TodoHandler {
	return &TodoHandler{
		todoService: todoService,
		tmpl:        tmpl,
		store:       store,
	}
}

func (h *TodoHandler) Add(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusInternalServerError)
		return
	}

	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get user session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)

	if user == nil {
		http.Error(w, "Usermust be logged in", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")

	todo, err := h.todoService.Create(user.ID, description)
	if err != nil {
		http.Error(w, "something went wrong whilecreating a new todo", http.StatusInternalServerError)
		return
	}

	err = h.tmpl.ExecuteTemplate(w, "todo", todo)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}

}
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {}
func (h *TodoHandler) Remove(w http.ResponseWriter, r *http.Request) {

	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	if user != nil {
		vars := mux.Vars(r)
		todoID := vars["id"]
		id, err := strconv.Atoi(todoID)
		if err != nil {
			http.Error(w, "path does not contain valid id", http.StatusBadRequest)
			return
		}

		todo, err := h.todoService.GetByID(id)
		if err != nil {
			http.Error(w, "could not get todo", http.StatusInternalServerError)
			return
		}

		userIsAuthor := user.ID == todo.UserID
		if userIsAuthor {
			err := h.todoService.Remove(todo.ID)
			if err != nil {
				http.Error(w, "could not remove todo", http.StatusInternalServerError)
				return
			}

		}

	}
}
