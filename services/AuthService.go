package services

import (
	"fmt"
	"go-todo/models"
	"go-todo/repositories"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(u *repositories.UserRepository) *AuthService {
	return &AuthService{
		userRepo: u,
	}
}

func (s *AuthService) Login(email string, password string) (*models.User, error) {
	userRecord, err := s.userRepo.GetUserRecordByEmail(email)
	if err != nil {
		return nil, err
	}

	if userRecord != nil {
		fmt.Println("user found", userRecord.Email)
	}

	if userRecord == nil {
		fmt.Println("user not found")
		return nil, fmt.Errorf("user not found")
	}

	if userRecord.Password != password {
		return nil, fmt.Errorf("incorrect password")
	}

	user := models.NewUser(userRecord.ID, userRecord.Name, userRecord.Email)

	return &user, nil
}
