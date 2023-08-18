package handlers

import (
	"go-todo/models"
	"html/template"
	"net/http"

	"github.com/michaeljs1990/sqlitestore"
)

type PageHandler struct {
	tmpl  *template.Template
	store *sqlitestore.SqliteStore
}

func NewPageHandler(tmpl *template.Template, store *sqlitestore.SqliteStore) *PageHandler {
	return &PageHandler{
		tmpl:  tmpl,
		store: store,
	}
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", 500)
	}
	user := GetUserFromSession(session)

	type HomePageData struct {
		User *models.User
	}

	var homePageData HomePageData = HomePageData{
		User: user,
	}
	err = h.tmpl.ExecuteTemplate(w, "home", homePageData)
	if err != nil {
		http.Error(w, "Could not render homepage", 500)
	}
}
