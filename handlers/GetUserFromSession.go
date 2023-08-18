package handlers

import (
	"go-todo/models"

	"github.com/gorilla/sessions"
)

func GetUserFromSession(s *sessions.Session) *models.User {
	val := s.Values["user"]
	var user = models.User{}
	user, ok := val.(models.User)
	if !ok {
		return nil
	}
	return &user
}
