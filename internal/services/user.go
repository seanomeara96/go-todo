package services

import (
	"fmt"
	"go-todo/internal/models"
	"html"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

type userLoginErrors struct {
	EmailErrors    []string
	PasswordErrors []string
}

func (s *Service) Login(email string, password string) (*models.User, *userLoginErrors, error) {
	var user models.User
	userLoginErrors := userLoginErrors{
		PasswordErrors: []string{},
		EmailErrors:    []string{},
	}

	isValidEmail := isValidEmail(email)
	if !isValidEmail {
		userLoginErrors.EmailErrors = append(userLoginErrors.EmailErrors, "You've provided an invalid email.")
	}

	userRecord, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get user by email. %w", err)
	}

	if userRecord == nil {
		userLoginErrors.EmailErrors = append(userLoginErrors.EmailErrors, "Could not find user with that email")
	}

	if len(userLoginErrors.EmailErrors) > 0 {
		return nil, &userLoginErrors, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(userRecord.Password), []byte(password))

	if err != nil {
		userLoginErrors.PasswordErrors = append(userLoginErrors.PasswordErrors, "Incorrect Password")
	}

	if len(userLoginErrors.PasswordErrors) > 0 {
		return nil, &userLoginErrors, nil
	}

	user = models.NewUser(userRecord.ID, userRecord.Name, userRecord.Email, "", userRecord.IsPaidUser, "")

	return &user, nil, nil
}

type userSignupErrors struct {
	UsernameErrors []string
	EmailErrors    []string
	PasswordErrors []string
}

// sign up for a new account
func (s *Service) NewUser(username, email, password string) (*models.User, *userSignupErrors, error) {
	userSignupErrors := userSignupErrors{[]string{}, []string{}, []string{}}

	// sanitize and  clean username, email & password
	id := uuid.New().String()

	username = html.EscapeString(username)
	email = html.EscapeString(email)

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if username == "" {
		userSignupErrors.UsernameErrors = append(userSignupErrors.UsernameErrors, "You must provide a user name.")
	}

	if email == "" {
		userSignupErrors.EmailErrors = append(userSignupErrors.EmailErrors, "You must provide an email.")
	}

	if password == "" {
		userSignupErrors.PasswordErrors = append(userSignupErrors.PasswordErrors, "You must provide a password.")
	}

	if username == "" || email == "" || password == "" {
		return nil, &userSignupErrors, nil
	}

	if !isValidEmail(email) && email != "" {
		userSignupErrors.EmailErrors = append(userSignupErrors.EmailErrors, "You must provide a valid email.")
	}

	// does email exist
	emailExists, err := s.repo.UserEmailExists(email)
	if err != nil {
		return nil, nil, err
	}

	if emailExists {
		userSignupErrors.EmailErrors = append(userSignupErrors.EmailErrors, "An account for this email already exists. Please Log In.")
	}

	// if any errors
	if len(userSignupErrors.EmailErrors) > 0 || len(userSignupErrors.UsernameErrors) > 0 || len(userSignupErrors.PasswordErrors) > 0 {
		return nil, &userSignupErrors, nil
	}

	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return nil, nil, err
	}

	userToInsert := models.NewUser(id, username, email, string(hashedpassword), false, "")
	err = s.repo.SaveUser(userToInsert)
	if err != nil {
		return nil, nil, err
	}

	user := models.NewUser(
		userToInsert.ID,
		userToInsert.Name,
		userToInsert.Email,
		"",
		userToInsert.IsPaidUser,
		"",
	)

	return &user, nil, nil
}

func (s *Service) UserCanCreateNewTodo(user *models.User, list []*models.Todo) (bool, error) {
	userIsPaidUser, internalErr := s.UserIsPaidUser(user.ID)
	if internalErr != nil {
		return false, fmt.Errorf("Could not determine payment status for user. %w", internalErr)
	}

	canCreateNewTodo := (!userIsPaidUser && len(list) < DefaultLimit) || userIsPaidUser

	return canCreateNewTodo, nil
}

func (s *Service) AddStripeIDToUser(userID, stripeID string) error {
	internalErr := s.repo.AddStripeIDToUser(userID, stripeID)
	if internalErr != nil {
		return fmt.Errorf("Could not add stripe id to user. %w", internalErr)
	}
	return nil
}

func (s *Service) GetUserByID(userID string) (*models.User, error) {
	user, internalErr := s.repo.GetUserByID(userID)
	if internalErr != nil {
		return nil, fmt.Errorf("Could not get user by ID. %w", internalErr)
	}
	return user, nil
}

func (s *Service) GetUserByEmail(email string) (*models.User, error) {
	user, internalErr := s.repo.GetUserByEmail(email)
	if internalErr != nil {
		return nil, fmt.Errorf("Could not get user by email. %w", internalErr)
	}
	return user, nil
}

func (s *Service) GetUserByStripeID(customerStripeID string) (*models.User, error) {
	user, err := s.repo.GetUserByStripeID(customerStripeID)
	if err != nil {
		return nil, fmt.Errorf("Could not get user by stripe ID. %w", err)
	}
	return user, nil

}

func (s *Service) UpdateUserPaymentStatus(userID string, isPaidUser bool) error {
	internalErr := s.repo.UpdateUserPaymentStatus(userID, isPaidUser)
	if internalErr != nil {
		return fmt.Errorf("Could not update user payment status. %w", internalErr)
	}

	return nil
}

func (s *Service) UserIsPaidUser(userID string) (bool, error) {
	isPaidUser, internalErr := s.repo.UserIsPaidUser(userID)
	if internalErr != nil {
		return false, fmt.Errorf("Could not determine whether user is paid user. %w", internalErr)
	}
	return isPaidUser, nil
}
