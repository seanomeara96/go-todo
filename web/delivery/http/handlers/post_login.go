package handlers

import (
	"fmt"
	"go-todo/internal/models"
	"go-todo/internal/server/renderer"
	"net/http"
)

// POST /login
/*
	Redirects to homepage when login is success. If there are client
	errors then the homepage will be rerendered with error messages
*/
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, USER_SESSION)
	user, err := h.getUserFromSession(session, err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user != nil {
		// user already logged in
		noCacheRedirect("/", w, r)
		return
	}

	/*
		no need to call logout if user == nil
		as we'll just render the login form instead.
	*/

	err = r.ParseForm()
	if err != nil {
		h.logger.Error("Could not parse login form")
		h.logger.Debug(err.Error())
		http.Error(w, "Could not parse form", http.StatusInternalServerError)
		return
	}

	email, password := r.FormValue("email"), r.FormValue("password")

	user, userErrors, err := h.service.Login(email, password)
	if err != nil {
		h.logger.Error("Could not log in user")
		h.logger.Debug(err.Error())
		http.Error(w, "error logging user in", http.StatusInternalServerError)
		return
	}

	if userErrors != nil {
		loginFormProps := renderer.NewLoginFormProps(userErrors.EmailErrors, userErrors.PasswordErrors)
		basePageProps := renderer.NewBasePageProps(nil)
		todoListProps := renderer.NewTodoListProps([]*models.Todo{}, false)
		homePageProps := renderer.NewHomePageProps(basePageProps, todoListProps, loginFormProps)
		bytes, err := h.render.HomePage(homePageProps)
		if err != nil {
			http.Error(w, "could not render homepage", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(bytes)

		if err != nil {
			h.logger.Error("Could not write homepage")
			h.logger.Debug(err.Error())
			http.Error(w, "could not write homepage", http.StatusInternalServerError)
		}

		return
	}

	session.Values["user"] = user.ID
	session.Save(r, w)
	noCacheRedirect("/", w, r)

	infoMsg := fmt.Sprintf("Session created for user (%s) logged in", user.ID)
	h.logger.Info(infoMsg)
}
