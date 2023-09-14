package handlers

import (
	"go-todo/renderer"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (h *Handler) userCanCreateNewTodo(user *models.User, list []*models.Todo)(bool, error){
	userIsPaidUser, err := h.service.UserIsPaidUser(user.ID)
	if err != nil {
		return false, err
	}
	return (!userIsPaidUser && len(list) < 10) || userIsPaidUser, nil
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
		http.Error(w, "Usermust be logged in", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")

	_, err = h.service.Create(user.ID, description)
	if err != nil {
		http.Error(w, "something went wrong whilecreating a new todo", http.StatusInternalServerError)
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
	if user != nil {
		noCacheRedirect(w, r)
		return
	}

	// get user from database
	user,err := h.service.GetUserByID(user.id)
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

	todo, err := h.service.GetByID(todoID)
	if err != nil {
		http.Error(w, "could not get todo", http.StatusInternalServerError)
		return
	}

	userIsNotAuthor := user.ID != todo.UserID
	if userIsNotAuthor {
		http.Error(w, "not authorized", http.StatusBadRequest)
		return
	} 
	
	err := h.service.Remove(todo.ID)
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

	if user != nil {
		vars := mux.Vars(r)
		idParam := vars["id"]
		todoID, err := strconv.Atoi(idParam)
		if err != nil {
			http.Error(w, "path does not contain valid id", http.StatusBadRequest)
			return
		}

		todo, err := h.service.UpdateStatus(user.ID, todoID)
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
