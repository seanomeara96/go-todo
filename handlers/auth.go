package handlers

import (
	"net/http"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
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

	user, err = h.service.Login(email, password)
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

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	err = h.store.Delete(r, w, session)
	if err != nil {
		http.Error(w, "could not delete session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
