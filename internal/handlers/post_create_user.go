package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {

		return err
	}

	name := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// TODO sanitize and clean input

	newUser, userErrors, err := h.service.NewUser(name, email, password)
	if err != nil {
		return err
	}

	if userErrors != nil {
		basePageProps := renderer.NewBasePageProps(nil)
		signUpFormProps := renderer.NewSignupFormProps(userErrors.UsernameErrors, userErrors.EmailErrors, userErrors.PasswordErrors)
		signupPageProps := renderer.NewSignupPageProps(basePageProps, signUpFormProps)
		signupPageBytes, err := h.render.Signup(signupPageProps)
		if err != nil {
			return err
		}
		if _, err := w.Write(signupPageBytes); err != nil {
			return err
		}
		return nil
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

	infoMsg := fmt.Sprintf("New user (%s) created", newUser.ID)
	h.logger.Info(infoMsg)

	return nil
}
