package handlers

import (
	"go-todo/models"
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

	_, err = h.todoService.Create(user.ID, description)
	if err != nil {
		http.Error(w, "something went wrong whilecreating a new todo", http.StatusInternalServerError)
		return
	}

	list, err := h.todoService.GetUserTodoList(user.ID)
	if err != nil {
		http.Error(w, "something went wrong while fetching todos", http.StatusInternalServerError)
		return
	}

	type TodoListParams struct {
		Todos            []*models.Todo
		TodoLimitReached bool
	}

	userIsPayedUser := false
	todoLimitReached := false
	if !userIsPayedUser && len(list) > 9 {
		todoLimitReached = true
	}

	data := TodoListParams{
		Todos:            list,
		TodoLimitReached: todoLimitReached,
	}

	err = h.tmpl.ExecuteTemplate(w, "todo-list", data)
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
			// TODO this is duplicate code
			list, err := h.todoService.GetUserTodoList(user.ID)
			if err != nil {
				http.Error(w, "something went wrong while fetching todos", http.StatusInternalServerError)
				return
			}

			type TodoListParams struct {
				Todos            []*models.Todo
				TodoLimitReached bool
			}

			userIsPayedUser := false
			todoLimitReached := false
			if !userIsPayedUser && len(list) > 9 {
				todoLimitReached = true
			}

			data := TodoListParams{
				Todos:            list,
				TodoLimitReached: todoLimitReached,
			}

			err = h.tmpl.ExecuteTemplate(w, "todo-list", data)
			if err != nil {
				http.Error(w, "could not render todo", http.StatusInternalServerError)
				return
			}

		} else {
			http.Error(w, "not authorized", http.StatusBadRequest)
			return
		}

	} else {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}
}

func (h *TodoHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)

	if user != nil {
		vars := mux.Vars(r)
		idParam := vars["id"]
		todoID, err := strconv.Atoi(idParam)
		if err != nil {
			http.Error(w, "path does not contain valid id", http.StatusBadRequest)
			return
		}

		todo, err := h.todoService.UpdateStatus(user.ID, todoID)
		if err != nil {
			http.Error(w, "could not update todo", http.StatusInternalServerError)
			return
		}

		if todo != nil {
			err = h.tmpl.ExecuteTemplate(w, "todo", todo)
			if err != nil {
				http.Error(w, "could not render todo", http.StatusInternalServerError)
				return
			}
			return
		}

		http.Error(w, "service did not return a todo item", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
