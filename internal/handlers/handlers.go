package handlers

import (
	"go-todo/internal/logger"
	"go-todo/internal/models"
	"go-todo/internal/server/renderer"
	"go-todo/internal/services"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

const STRIPE_API_KEY = "STRIPE_API_KEY"
const USER_SESSION = "user-session"
const STRIPE_WEBHOOK_SECRET = "STRIPE_WEBHOOK_SECRET"

type Handler struct {
	service *services.Service
	store   *sqlitestore.SqliteStore
	render  *renderer.Renderer
	logger  *logger.Logger
}

func NewHandler(service *services.Service, store *sqlitestore.SqliteStore, renderer *renderer.Renderer, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		store:   store,
		render:  renderer,
		logger:  logger,
	}
}

func (h *Handler) getUserFromSession(s *sessions.Session, err error) (*models.User, error) {
	if err != nil {
		return nil, err
	}
	val := s.Values["user"]
	userID, ok := val.(string)
	if !ok {
		return nil, nil
	}
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func noCacheRedirect(path string, w http.ResponseWriter, r *http.Request) {
	// Set cache-control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Redirect the user to a new URL
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// func (h *Handler) Upgrade(w http.ResponseWriter, r *http.Request) {

// 	// stripe code

// }
