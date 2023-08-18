package handlers

import (
	"go-todo/services"
	"net/http"
)

type TodoHandler struct {
	t *services.TodoService
}

func NewTodoHandler(t *services.TodoService) *TodoHandler {
	return &TodoHandler{
		t: t,
	}
}

func (*TodoHandler) Add(w http.ResponseWriter, r *http.Request)    {}
func (*TodoHandler) Remove(w http.ResponseWriter, r *http.Request) {}
