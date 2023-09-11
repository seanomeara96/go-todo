package services

import (
	"go-todo/models"

	"github.com/google/uuid"
)

func (s *Service) NewUser(username, email, password string) (*models.User, error) {
	// sanitize and  clean usernalme email
	id := uuid.New().String()

	// need to remomber to hash password first
	userToInsert := models.NewUserRecord(id, username, email, password, false)
	err := s.repo.SaveUser(userToInsert)
	if err != nil {
		return nil, err
	}
	user := models.NewUser(userToInsert.ID, userToInsert.Email, userToInsert.Name, userToInsert.IsPaidUser)
	return &user, nil
}

func (s *Service) UserIsPayedUser(userID string) bool {
	return false
}

func (s *Service) AddStripeIDToUser(userID, stripeID string) error {
	return s.repo.AddStripeIDToUser(userID, stripeID)
}

func (s *Service) GetUserByEmail(email string) (*models.User, error) {
	return s.repo.GetUserByEmail(email)
}
