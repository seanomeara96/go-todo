package services

import (
	"fmt"
	"go-todo/models"
	"go-todo/repositories"
	"html"
	"regexp"

	"github.com/google/uuid"
)

type Service struct {
	repo *repositories.Repository
}

func NewService(r *repositories.Repository) *Service {
	return &Service{
		repo: r,
	}
}

type userLoginErrors struct {
	EmailErrors    []string
	PasswordErrors []string
}

func (s *Service) Login(email string, password string) (*models.User, *userLoginErrors, error) {
	var user models.User
	var EmailErrors []string
	var PasswordErrors []string

	userRecord, err := s.repo.GetUserRecordByEmail(email)
	if err != nil {
		return nil, nil, err
	}

	if userRecord == nil {
		EmailErrors = append(EmailErrors, "Could not find user with that email")

		userLoginErrors := userLoginErrors{
			PasswordErrors: PasswordErrors,
			EmailErrors:    EmailErrors,
		}

		return nil, &userLoginErrors, nil
	}

	if userRecord.Password != password {
		PasswordErrors = append(PasswordErrors, "Incorrect Password")

		userLoginErrors := userLoginErrors{
			PasswordErrors: PasswordErrors,
			EmailErrors:    EmailErrors,
		}

		return nil, &userLoginErrors, nil
	}

	user = models.NewUser(userRecord.ID, userRecord.Name, userRecord.Email, userRecord.IsPaidUser)

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

	// need to remember to hash password first
	userToInsert := models.NewUserRecord(id, username, email, password, false)
	err = s.repo.SaveUser(userToInsert)
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
