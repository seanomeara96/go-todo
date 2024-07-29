package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (h *Handler) UpdateTodoStatus(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
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

	todo, clientError, err := h.service.UpdateTodoStatus(user.ID, todoID)
	if err != nil {
		http.Error(w, "could not update todo", http.StatusInternalServerError)
		return
	}

	if clientError != nil {
		http.Error(w, clientError.Message, clientError.Code)
		return
	}

	if todo == nil {
		http.Error(w, "service did not return a todo item", http.StatusInternalServerError)
		return
	}

	todoBytes, err := h.render.Todo(todo)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}

	w.Write(todoBytes)

	infoMsg := fmt.Sprintf("User (%s) updated todo (%d) to status: %v", user.ID, todo.ID, todo.IsComplete)
	h.logger.Info(infoMsg)
}
