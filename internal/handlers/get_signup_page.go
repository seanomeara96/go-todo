package handlers

import (
	"go-todo/internal/models"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) SignupPage(w http.ResponseWriter, r *http.Request) error {
	user, _ := r.Context().Value(userIDKey).(*models.User)
	if user != nil {
		return noCacheRedirect("/", w, r)
	}

	basePageProps := renderer.NewBasePageProps(user)
	noErrors := []string{}
	signupFormProps := renderer.NewSignupFormProps(noErrors, noErrors, noErrors)
	signupPageProps := renderer.NewSignupPageProps(basePageProps, signupFormProps)
	signupPageBytes, err := h.render.Signup(signupPageProps)
	if err != nil {
		return err
	}

	_, err = w.Write(signupPageBytes)
	if err != nil {
		return err
	}

	return nil
}
