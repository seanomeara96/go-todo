package services

import (
	"go-todo/models"
	"go-todo/repositories"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) NewUser(username, email, password string) (*models.User, error) {
	// sanitize and  clean usernalme email
	id := uuid.New().String()

	// need to remomber to hash password first
	userToInsert := models.NewUserRecord(id, username, email, password)
	err := s.userRepo.Save(userToInsert)
	if err != nil {
		return nil, err
	}
	user := models.NewUser(userToInsert.ID, userToInsert.Email, userToInsert.Name)
	return &user, nil
}

func (s *UserService) UserIsPayedUser(userID string) bool {
	return false
}
