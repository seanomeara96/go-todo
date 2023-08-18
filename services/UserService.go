package services

import (
	"go-todo/models"
	"go-todo/repositories"
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
	// need to remomber to hash password first
	userToInsert := models.NewUserRecord(username, email, password)
	err := s.userRepo.Save(userToInsert)
	if err != nil {
		nil, err
	}
	user := models.NewUser(userTo)
}
