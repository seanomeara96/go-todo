package handlers

import (
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) SignupPage(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		return err
	}

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
