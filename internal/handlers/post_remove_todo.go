package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

func (h *Handler) RemoveTodo(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		return err
	}

	if user == nil {
		return h.Logout(w, r)
	}

	todoID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return err
	}

	todo, clientError, internalError := h.service.GetTodoByID(todoID, user.ID)
	if internalError != nil {
		return err
	}

	if clientError != nil {
		return fmt.Errorf("client err", clientError.Message)
	}

	clientError, internalError = h.service.DeleteTodo(todo.ID, user.ID)
	if internalError != nil {
		return err
	}

	if clientError != nil {
		return fmt.Errorf("%s:%d", clientError.Message, clientError.Code)
	}

	if _, err := w.Write([]byte("")); err != nil {
		return err
	}

	infoMsg := fmt.Sprintf("User (%s) removed todo (%s)", user.ID, r.PathValue("id"))
	h.logger.Info(infoMsg)

	return nil
}
