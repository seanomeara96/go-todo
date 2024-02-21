package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) AddTodo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusInternalServerError)
		return
	}

	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	// TODO render client errors
	lastInsertedTodo, _, err := h.service.CreateTodo(user.ID, r.FormValue("description"))
	// TODO for now i am returning an error if todo is nil
	if err != nil || lastInsertedTodo == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := h.service.GetUserTodoList(user.ID)
	if err != nil {
		http.Error(w, "something went wrong while fetching todos", http.StatusInternalServerError)
		return
	}

	canCreateNewTodo, err := h.service.UserCanCreateNewTodo(user, list)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
	todoList, err := h.render.TodoList(todoListProps)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}

	w.Write(todoList)

	infoMsg := fmt.Sprintf("User (%s) added a todo (%d)", user.ID, lastInsertedTodo.ID)
	h.logger.Info(infoMsg)
}
