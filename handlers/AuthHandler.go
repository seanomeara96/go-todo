package handlers

import (
	"go-todo/services"
	"net/http"

	"github.com/michaeljs1990/sqlitestore"
)

type AuthHandler struct {
	auth  *services.AuthService
	store *sqlitestore.SqliteStore
}

func NewAuthHandler(service *services.AuthService, store *sqlitestore.SqliteStore) *AuthHandler {
	return &AuthHandler{
		auth:  service,
		store: store,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		panic(err)
	}

	user := GetUserFromSession(session)

	if user != nil {
		// user already logged in
		http.Redirect(w, r, "", http.StatusFound)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err = h.auth.Login(email, password)
	if err != nil {
		http.Error(w, "Something went wrong during login", http.StatusInternalServerError)
		return
	}

	if user == nil {

		http.Error(w, "Incorrect credentials", http.StatusBadRequest)
		return
	}

	session.Values["user"] = *user
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "user-session")
	err := h.store.Delete(r, w, session)
	if err != nil {
		http.Error(w, "could not delete session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
