package services

import (
	"regexp"
	"go-todo/models"
	
	"github.com/google/uuid"
)

func isValidEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

func (s *Service) NewUser(username, email, password string) (*models.User, error) {
	// sanitize and  clean usernalme email
	id := uuid.New().String()

	username = html.EscapeString(username)
	email = html.EscapeString(email)

	if !isValidEmail(email) {
		return nil, fmt.Errorf("must provide a valid email")
	}

	// does email exist
	found, err := s.repo.UserEmailExists(email)
	if err != nil {
		return nil, fmt.Errorf("could not determin existence of email")
	}

	if found {
		return nil, fmt.Errorf("must supply unique email")
	}

	// need to remomber to hash password first
	userToInsert := models.NewUserRecord(id, username, email, password, false)
	err := s.repo.SaveUser(userToInsert)
	if err != nil {
		return nil, err
	}

	user := models.NewUser(
		userToInsert.ID, 
		userToInsert.Email, 
		userToInsert.Name, 
		userToInsert.IsPaidUser,
	)

	return &user, nil
}

func (s *Service) AddStripeIDToUser(userID, stripeID string) error {
	return s.repo.AddStripeIDToUser(userID, stripeID)
}

func (s *Service) GetUserByEmail(email string) (*models.User, error) {
	return s.repo.GetUserByEmail(email)
}

func (s *Service) UpdateUserPaymentStatus(userID string, isPaidUser bool) error {
	return s.repo.UpdateUserPaymentStatus(userID, isPaidUser)
}

func (s *Service) UserIsPaidUser(userID string) (bool, error) {
	return s.repo.UserIsPaidUser(userID)
}
