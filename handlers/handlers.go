package handlers

import (
	"go-todo/models"
	"go-todo/renderer"
	"go-todo/services"

	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

type Handler struct {
	service  *services.Service
	store    *sqlitestore.SqliteStore
	renderer *renderer.Renderer
}

func NewHandler(service *services.Service, store *sqlitestore.SqliteStore, renderer *renderer.Renderer) *Handler {
	return &Handler{
		service:  service,
		store:    store,
		renderer: renderer,
	}
}

func GetUserFromSession(s *sessions.Session) *models.User {
	val := s.Values["user"]
	var user = models.User{}
	user, ok := val.(models.User)
	if !ok {
		return nil
	}
	return &user
}
