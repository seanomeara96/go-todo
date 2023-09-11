package services

import (
	"fmt"
	"go-todo/models"
)

func (s *Service) Login(email string, password string) (*models.User, error) {
	userRecord, err := s.repo.GetUserRecordByEmail(email)
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

	user := models.NewUser(userRecord.ID, userRecord.Name, userRecord.Email, userRecord.IsPaidUser)

	return &user, nil
}
