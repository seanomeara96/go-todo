package handlers

import (
	"go-todo/models"
	"go-todo/renderer"
	"go-todo/services"
	"net/http"

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

func noCacheRedirect(w http.ResponseWriter, r *http.Request) {
	// Set cache-control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Redirect the user to a new URL
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
