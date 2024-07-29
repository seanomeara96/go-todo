package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

func (h *Handler) UpdateTodoStatus(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUserFromContext(r)
	if err != nil {
		return err
	}

	if user == nil {
		return h.Logout(w, r)
	}

	user, err = h.service.GetUserByID(user.ID)
	if err != nil {
		return err
	}

	todoID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("path does not contain valid id %d", http.StatusBadRequest)
	}

	todo, clientError, err := h.service.UpdateTodoStatus(user.ID, todoID)
	if err != nil {
		return err
	}

	if clientError != nil {
		return fmt.Errorf("%s:%d", clientError.Message, clientError.Code)
	}

	if todo == nil {
		return fmt.Errorf("service did not return a todo item")
	}

	todoBytes, err := h.render.Todo(todo)
	if err != nil {
		return err
	}

	if _, err := w.Write(todoBytes); err != nil {
		return err
	}

	infoMsg := fmt.Sprintf("User (%s) updated todo (%d) to status: %v", user.ID, todo.ID, todo.IsComplete)
	h.logger.Info(infoMsg)
	return nil
}
