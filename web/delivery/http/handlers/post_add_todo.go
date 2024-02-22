package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) AddTodo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Could not parse form at add todo")
		h.logger.Debug(fmt.Sprintf("Error parsing form at addtodo. %v", err))
		w.WriteHeader(500)
		return
	}

	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		w.WriteHeader(500)
		h.logger.Error("Could not get user from session")
		h.logger.Debug(fmt.Sprintf("Error fetching user from session. %v", err))
		return
	}

	// user session but no user
	if user == nil {
		h.Logout(w, r)
		return
	}

	// TODO render client errors
	_, clientErrors, err := h.service.CreateTodo(user.ID, r.FormValue("description"))
	// TODO for now i am returning an error if todo is nil
	if err != nil {
		w.WriteHeader(500)
		h.logger.Error("Could not create todo at add todo")
		h.logger.Debug(fmt.Sprintf("Error creating todo at add todo. %v", err))
		return
	}

	list, err := h.service.GetUserTodoList(user.ID)
	if err != nil {
		h.logger.Error("Could not get user todo list at add todo")
		h.logger.Debug(fmt.Sprintf("Error getting todo list at add todo. %v", err))
		w.WriteHeader(500)
		return
	}

	canCreateNewTodo, err := h.service.UserCanCreateNewTodo(user, list)
	if err != nil {
		h.logger.Error("Cant determine whether user can create new todo at add todo")
		h.logger.Debug(fmt.Sprintf("Error determining whether user can create new todo %v", err))
		w.WriteHeader(500)
		return
	}

	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo, clientErrors)
	todoList, err := h.render.TodoList(todoListProps)
	if err != nil {
		w.WriteHeader(500)
		h.logger.Error("Cannot render todo list")
		h.logger.Debug(fmt.Sprintf("Error rendering todo list at add todo. %v", err))
		return
	}

	w.Write(todoList)

	if clientErrors == nil {
		infoMsg := fmt.Sprintf("User (%s) added a todo", user.ID)
		h.logger.Info(infoMsg)

	}

}
