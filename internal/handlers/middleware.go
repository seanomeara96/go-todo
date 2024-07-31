package handlers

import (
	"context"
	"fmt"
	"go-todo/internal/models"
	"net/http"
)

type key string

const userIDKey key = "user"

func (h *Handler) PathLogger(next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		h.logger.Info(fmt.Sprintf("%s => %s", r.Method, r.URL.Path))
		return next(w, r)
	}
}

func (h *Handler) UserMustBeLoggedIn(next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// do something

		user, ok := r.Context().Value(userIDKey).(*models.User)
		if !ok || user == nil {
			return fmt.Errorf("user must be logged in")
		}

		return next(w, r)
	}
}

func (h *Handler) AddUserToContext(next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
		if err != nil {
			return fmt.Errorf("could not get user from session in middleware, %w", err)
		}

		ctx := context.WithValue(r.Context(), userIDKey, user)

		return next(w, r.WithContext(ctx))
	}
}
