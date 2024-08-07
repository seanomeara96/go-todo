package handlers

import (
	"errors"
	"fmt"
	"go-todo/internal/logger"
	"go-todo/internal/models"
	"go-todo/internal/server/renderer"
	"go-todo/internal/services"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

type Handler struct {
	service *services.Service
	store   *sqlitestore.SqliteStore
	render  *renderer.Renderer
	logger  *logger.Logger
}

type HandleFunc func(w http.ResponseWriter, r *http.Request) error
type MiddleWareFunc func(next HandleFunc) HandleFunc

const STRIPE_API_KEY = "STRIPE_API_KEY"
const USER_SESSION = "user-session"
const STRIPE_WEBHOOK_SECRET = "STRIPE_WEBHOOK_SECRET"

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

func (h *Handler) getUserFromContext(r *http.Request) (*models.User, error) {
	user, ok := r.Context().Value(userIDKey).(*models.User)
	if !ok {
		return nil, fmt.Errorf("coul not get user pointer from contex")
	}
	return user, nil
}

func noCacheRedirect(path string, w http.ResponseWriter, r *http.Request) error {
	// Set cache-control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Redirect the user to a new URL
	http.Redirect(w, r, path, http.StatusSeeOther)
	return nil
}

// func (h *Handler) Upgrade(w http.ResponseWriter, r *http.Request) {

// 	// stripe code

// }

func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) error {
	return errors.New("not implemented")
}

func (h *Handler) AnalyticsDashboard(w http.ResponseWriter, r *http.Request) error {
	return errors.New("not implemented")
}

func (h *Handler) UsersPage(w http.ResponseWriter, r *http.Request) error {
	// h.service.GetUsers()
	return errors.New("not implemented")
}

func (h *Handler) UserProfilePage(w http.ResponseWriter, r *http.Request) error {
	_, err := h.service.GetUserByID(r.PathValue("user_id"))
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	// h.service.GetUsers()
	return errors.New("not implemented")
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	// h.service.GetUsers()
	return errors.New("not implemented")
}
