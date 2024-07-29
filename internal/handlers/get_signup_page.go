package handlers

import (
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) SignupPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user != nil {
		noCacheRedirect("/", w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	noErrors := []string{}
	signupFormProps := renderer.NewSignupFormProps(noErrors, noErrors, noErrors)
	signupPageProps := renderer.NewSignupPageProps(basePageProps, signupFormProps)
	signupPageBytes, err := h.render.Signup(signupPageProps)
	if err != nil {
		h.logger.Error("could not render signup page")
		h.logger.Debug(err.Error())
		http.Error(w, "could not render sigup page", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(signupPageBytes)
	if err != nil {
		h.logger.Error("Could not write signup page")
		h.logger.Debug(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
