package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) AddTodo(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("Error parsing form at addtodo. %v", err)
	}

	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		return fmt.Errorf("error fetching user from session. %v", err)
	}

	// user session but no user
	if user == nil {
		return h.Logout(w, r)
	}

	// TODO render client errors
	_, clientErrors, err := h.service.CreateTodo(user.ID, r.FormValue("description"))
	// TODO for now i am returning an error if todo is nil
	if err != nil {
		return err
	}

	list, err := h.service.GetUserTodoList(user.ID)
	if err != nil {
		return fmt.Errorf("Error getting todo list at add todo. %v", err)
	}

	canCreateNewTodo, err := h.service.UserCanCreateNewTodo(user, list)
	if err != nil {
		return (fmt.Errorf("Error determining whether user can create new todo %v", err))
	}

	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo, clientErrors)
	todoList, err := h.render.TodoList(todoListProps)
	if err != nil {
		return err
	}

	if _, err := w.Write(todoList); err != nil {
		return err
	}

	if clientErrors == nil {
		infoMsg := fmt.Sprintf("User (%s) added a todo", user.ID)
		h.logger.Info(infoMsg)
	}

	return nil
}
