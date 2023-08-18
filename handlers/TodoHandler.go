package handlers

import (
	"go-todo/services"
	"net/http"

	"github.com/michaeljs1990/sqlitestore"
)

type TodoHandler struct {
	t     *services.TodoService
	store *sqlitestore.SqliteStore
}

func NewTodoHandler(t *services.TodoService, store *sqlitestore.SqliteStore) *TodoHandler {
	return &TodoHandler{
		t:     t,
		store: store,
	}
}

func (h *TodoHandler) Add(w http.ResponseWriter, r *http.Request) {
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

}
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {}
func (h *TodoHandler) Remove(w http.ResponseWriter, r *http.Request) {}
