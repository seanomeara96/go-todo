package services

import (
	"fmt"
	"go-todo/logger"
	"go-todo/models"
	"go-todo/repositories"
	"html"
	"log"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo   *repositories.Repository
	logger *logger.Logger
}

func NewService(r *repositories.Repository, logger *logger.Logger) *Service {
	return &Service{
		repo:   r,
		logger: logger,
	}
}

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

	userRecord, err := s.repo.GetUserRecordByEmail(email)
	if err != nil {
		return nil, nil, err
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

	user = models.NewUser(userRecord.ID, userRecord.Name, userRecord.Email, userRecord.IsPaidUser)
	log.Printf("User (%s) logged in successfully", user.Email)
	return &user, nil, nil
}

func (s *Service) CreateTodo(userID, description string) (*models.Todo, error) {
	if description == "" {
		return nil, fmt.Errorf("cannot supply an empty description")
	}

	sanitizedDescription := html.EscapeString(description)

	todo := models.NewTodo(userID, sanitizedDescription)

	lastInsertedTodoID, err := s.repo.CreateTodo(&todo)
	if err != nil {
		return nil, err
	}

	todo.ID = lastInsertedTodoID
	return &todo, nil
}

func (s *Service) GetUserTodoList(userID string) ([]*models.Todo, error) {
	return s.repo.GetAllTodosByUserID(userID)
}

func (s *Service) GetTodoByID(ID int) (*models.Todo, error) {
	return s.repo.GetTodoByID(ID)
}

func (s *Service) DeleteTodo(ID int) error {
	// TODO run auth check

	return s.repo.DeleteTodo(ID)
}

func (s *Service) DeleteAllTodosByUserID(userID string) error {
	return s.repo.DeleteAllTodosByUserID(userID)
}

func (s *Service) DeleteAllTodosByUserIDAndStatus(userID string, IsComplete bool) error {
	return s.repo.DeleteAllTodosByUserIDAndStatus(userID, IsComplete)
}

func (s *Service) UpdateTodoStatus(userID string, todoID int) (*models.Todo, error) {
	todo, err := s.repo.GetTodoByID(todoID)
	if err != nil {
		return nil, err
	}

	if todo == nil {
		return nil, err
	}

	userIsAuthor := userID == todo.UserID
	if !userIsAuthor {
		return nil, fmt.Errorf("not authorized")
	}

	updatedStatus := !todo.IsComplete

	todo.IsComplete = updatedStatus

	err = s.repo.UpdateTodo(*todo)
	if err != nil {
		return nil, err
	}

	return todo, nil
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

	userToInsert := models.NewUserRecord(id, username, email, string(hashedpassword), false)
	err = s.repo.SaveUser(userToInsert)
	if err != nil {
		return nil, nil, err
	}

	user := models.NewUser(
		userToInsert.ID,
		userToInsert.Name,
		userToInsert.Email,
		userToInsert.IsPaidUser,
	)

	log.Printf("User (%s) created successfully", user.Email)
	return &user, nil, nil
}

func (s *Service) UserCanCreateNewTodo(user *models.User, list []*models.Todo) (bool, error) {
	userIsPaidUser, err := s.UserIsPaidUser(user.ID)
	if err != nil {
		return false, err
	}

	canCreateNewTodo := (!userIsPaidUser && len(list) < 10) || userIsPaidUser

	return canCreateNewTodo, nil
}

func (s *Service) AddStripeIDToUser(userID, stripeID string) error {
	return s.repo.AddStripeIDToUser(userID, stripeID)
}

func (s *Service) GetUserByID(userID string) (*models.User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *Service) GetUserByEmail(email string) (*models.User, error) {
	return s.repo.GetUserByEmail(email)
}

func (s *Service) GetUserByStripeID(customerStripeID string) (*models.User, error) {
	return s.repo.GetUserByStripeID(customerStripeID)
}

func (s *Service) UpdateUserPaymentStatus(userID string, isPaidUser bool) error {
	return s.repo.UpdateUserPaymentStatus(userID, isPaidUser)
}

func (s *Service) UserIsPaidUser(userID string) (bool, error) {
	return s.repo.UserIsPaidUser(userID)
}
