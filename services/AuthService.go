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

func (a *AuthService) Login(email string, password string) (*models.User, error) {
	if email == "email" && password == "password" {
		user := models.NewUser(email, "username")
		return &user, nil
	}
	return nil, fmt.Errorf("something went wrong woth login creddentials")
}
