package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse signup form", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// TODO sanitize and clean input

	newUser, userErrors, err := h.service.NewUser(name, email, password)
	if userErrors != nil {
		basePageProps := renderer.NewBasePageProps(nil)
		signUpFormProps := renderer.NewSignupFormProps(userErrors.UsernameErrors, userErrors.EmailErrors, userErrors.PasswordErrors)
		signupPageProps := renderer.NewSignupPageProps(basePageProps, signUpFormProps)
		signupPageBytes, err := h.render.Signup(signupPageProps)
		if err != nil {
			http.Error(w, "could not render sigup page", http.StatusInternalServerError)
			return
		}
		w.Write(signupPageBytes)
		return
	}

	if err != nil {
		http.Error(w, "could not create user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

	infoMsg := fmt.Sprintf("New user (%s) created", newUser.ID)
	h.logger.Info(infoMsg)
}
