package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (h *Handler) RemoveTodo(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	vars := mux.Vars(r)
	todoIDString := vars["id"]

	todoID, err := strconv.Atoi(todoIDString)
	if err != nil {
		http.Error(w, "path does not contain valid id", http.StatusBadRequest)
		return
	}

	todo, clientError, internalError := h.service.GetTodoByID(todoID, user.ID)
	if internalError != nil {
		http.Error(w, "could not get todo", http.StatusInternalServerError)
		return
	}

	if clientError != nil {
		http.Error(w, clientError.Message, clientError.Code)
		return
	}

	clientError, internalError = h.service.DeleteTodo(todo.ID, user.ID)
	if internalError != nil {
		http.Error(w, "could not remove todo", http.StatusInternalServerError)
		return
	}

	if clientError != nil {
		http.Error(w, clientError.Message, clientError.Code)
		return
	}

	w.Write([]byte(""))

	infoMsg := fmt.Sprintf("User (%s) removed todo (%s)", user.ID, todoIDString)
	h.logger.Info(infoMsg)
}
