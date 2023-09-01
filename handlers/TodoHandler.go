package handlers

import (
	"go-todo/renderer"
	"go-todo/services"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/michaeljs1990/sqlitestore"
)

type TodoHandler struct {
	todoService *services.TodoService
	userService *services.UserService
	store       *sqlitestore.SqliteStore
	renderer    *renderer.Renderer
}

func NewTodoHandler(todoService *services.TodoService, userService *services.UserService, renderer *renderer.Renderer, store *sqlitestore.SqliteStore) *TodoHandler {
	return &TodoHandler{
		todoService: todoService,
		userService: userService,
		renderer:    renderer,
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

	userIsPayedUser := h.userService.UserIsPayedUser(user.ID)
	canCreateNewTodo := (!userIsPayedUser && len(list) < 10) || userIsPayedUser

	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
	todoList, err := h.renderer.TodoList(todoListProps)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}
	w.Write(todoList)
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

			userIsPayedUser := h.userService.UserIsPayedUser(user.ID)
			canCreateNewTodo := (!userIsPayedUser && len(list) < 10) || userIsPayedUser

			props := renderer.NewTodoListProps(list, canCreateNewTodo)
			todoListBytes, err := h.renderer.TodoList(props)
			if err != nil {
				http.Error(w, "could not render todo", http.StatusInternalServerError)
				return
			}
			w.Write(todoListBytes)
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
			todoBytes, err := h.renderer.Todo(todo)
			if err != nil {
				http.Error(w, "could not render todo", http.StatusInternalServerError)
				return
			}
			w.Write(todoBytes)
			return
		}

		http.Error(w, "service did not return a todo item", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
