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
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	session, err := h.store.Get(r, USER_SESSION)
	user, err := h.getUserFromSession(session, err)
	if err != nil {
		return err
	}

	if user != nil {
		// user already logged in
		return noCacheRedirect("/", w, r)

	}

	/*
		no need to call logout if user == nil
		as we'll just render the login form instead.
	*/

	err = r.ParseForm()
	if err != nil {
		return err
	}

	email, password := r.FormValue("email"), r.FormValue("password")

	user, userErrors, err := h.service.Login(email, password)
	if err != nil {
		return err
	}

	if userErrors != nil {
		loginFormProps := renderer.NewLoginFormProps(userErrors.EmailErrors, userErrors.PasswordErrors)
		basePageProps := renderer.NewBasePageProps(nil)
		todoListProps := renderer.NewTodoListProps([]*models.Todo{}, false, nil)
		homePageProps := renderer.NewHomePageProps(basePageProps, todoListProps, loginFormProps)
		bytes, err := h.render.HomePage(homePageProps)
		if err != nil {
			return err
		}

		_, err = w.Write(bytes)
		if err != nil {
			return err
		}

		return nil
	}

	session.Values["user"] = user.ID
	session.Save(r, w)

	infoMsg := fmt.Sprintf("Session created for user (%s) logged in", user.ID)
	h.logger.Info(infoMsg)

	return noCacheRedirect("/", w, r)
}
