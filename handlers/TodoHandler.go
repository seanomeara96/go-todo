package handlers

import (
	"go-todo/models"
	"go-todo/renderer"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (h *Handler) userCanCreateNewTodo(user *models.User, list []*models.Todo) (bool, error) {
	userIsPaidUser, err := h.service.UserIsPaidUser(user.ID)
	if err != nil {
		return false, err
	}

	canCreateNewTodo := (!userIsPaidUser && len(list) < 10) || userIsPaidUser

	return canCreateNewTodo, nil
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Usermust be logged in", http.StatusForbidden)
		return
	}

	_, err = h.service.CreateTodo(user.ID, r.FormValue("description"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := h.service.GetUserTodoList(user.ID)
	if err != nil {
		http.Error(w, "something went wrong while fetching todos", http.StatusInternalServerError)
		return
	}

	userIsPayedUser, err := h.service.UserIsPaidUser(user.ID)
	if err != nil {
		http.Error(w, "error  determingin user paymnet status", http.StatusInternalServerError)
		return
	}
	canCreateNewTodo := (!userIsPayedUser && len(list) < 10) || userIsPayedUser

	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
	todoList, err := h.renderer.TodoList(todoListProps)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}
	w.Write(todoList)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {}
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {

	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	if user == nil {
		http.Error(w, "user must be logged in", http.StatusForbidden)
		return
	}

	// get user from database
	user, err = h.service.GetUserByID(user.ID)
	if err != nil {
		http.Error(w, "could not find user", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	todoIDString := vars["id"]

	todoID, err := strconv.Atoi(todoIDString)
	if err != nil {
		http.Error(w, "path does not contain valid id", http.StatusBadRequest)
		return
	}

	todo, err := h.service.GetTodoByID(todoID)
	if err != nil {
		http.Error(w, "could not get todo", http.StatusInternalServerError)
		return
	}

	userIsNotAuthor := user.ID != todo.UserID
	if userIsNotAuthor {
		http.Error(w, "not authorized", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteTodo(todo.ID)
	if err != nil {
		http.Error(w, "could not remove todo", http.StatusInternalServerError)
		return
	}

	// TODO this is duplicate code
	list, err := h.service.GetUserTodoList(user.ID)
	if err != nil {
		http.Error(w, "something went wrong while fetching todos", http.StatusInternalServerError)
		return
	}

	userCanCreateNewTodo, err := h.userCanCreateNewTodo(user, list)
	if err != nil {
		http.Error(w, "error determining user payment status", http.StatusInternalServerError)
		return
	}

	props := renderer.NewTodoListProps(list, userCanCreateNewTodo)
	todoListBytes, err := h.renderer.TodoList(props)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}
	w.Write(todoListBytes)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)

	if user == nil {
		http.Error(w, "user must be logged in", http.StatusForbidden)
		return
	}

	user, err = h.service.GetUserByID(user.ID)
	if err != nil {
		http.Error(w, "trouble finding that user", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	idParam := vars["id"]
	todoID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "path does not contain valid id", http.StatusBadRequest)
		return
	}

	todo, err := h.service.UpdateTodoStatus(user.ID, todoID)
	if err != nil {
		http.Error(w, "could not update todo", http.StatusInternalServerError)
		return
	}

	if todo == nil {
		http.Error(w, "service did not return a todo item", http.StatusInternalServerError)
		return
	}

	todoBytes, err := h.renderer.Todo(todo)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}

	w.Write(todoBytes)
}
