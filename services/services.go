package services

import (
	"fmt"
	"go-todo/cache"
	"go-todo/logger"
	"go-todo/models"
	"go-todo/repositories"
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const DefaultLimit = 10

type internalError error
type clientError *ClientError

type Service struct {
	repo   *repositories.Repository
	caches *cache.Caches
	logger *logger.Logger
}

func NewService(r *repositories.Repository, caches *cache.Caches, logger *logger.Logger) *Service {
	return &Service{
		repo:   r,
		caches: caches,
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

	userRecord, err := s.repo.GetUserByEmail(email)
	if err != nil {
		s.logger.Error("Could not get user record by email")
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

	user = models.NewUser(userRecord.ID, userRecord.Name, userRecord.Email, "", userRecord.IsPaidUser, "")

	infoMsg := fmt.Sprintf("User (%s) successfully authenticated", userRecord.ID)
	s.logger.Info(infoMsg)

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
		s.logger.Error("Could not create new Todo item")
		return nil, err
	}

	todo.ID = lastInsertedTodoID

	infoMsg := fmt.Sprintf("User (%s) successfully created new todo (%d)", userID, todo.ID)
	s.logger.Info(infoMsg)

	return &todo, nil
}

func (s *Service) GetUserTodoList(userID string) ([]*models.Todo, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		errMsg := fmt.Sprintf("Something went wrong look for user by ID (%s)", user.ID)
		s.logger.Error(errMsg)
		return nil, err
	}

	if user == nil {
		errMsg := fmt.Sprintf("Could not find user (%s)", user.ID)
		s.logger.Error(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	limit := DefaultLimit
	if user.IsPaidUser {
		limit = 0
	}

	todoList, err := s.repo.GetTodosByUserID(userID, limit)
	if err != nil {
		errMsg := fmt.Sprintf("Could not get todos for user ID (%s)", userID)
		s.logger.Error(errMsg)
		return nil, err
	}

	infoMsg := fmt.Sprintf("User (%s) successfully retrieved  their todo list", userID)
	s.logger.Info(infoMsg)

	return todoList, nil
}

func (s *Service) GetTodoByID(ID int, userID string) (*models.Todo, clientError, internalError) {
	todo, err := s.repo.GetTodoByID(ID)
	// internal server error
	if err != nil {
		errMsg := fmt.Sprintf("Could not get todo id %d", ID)
		s.logger.Error(errMsg)
		return nil, nil, err
	}

	// client error
	if todo.UserID != userID {
		warningMsg := fmt.Sprintf("User (%s) attempted unauthorized access of resource", userID)
		s.logger.Warning(warningMsg)
		return nil, NewClientError("User not authorized", http.StatusUnauthorized), nil
	}

	infoMsg := fmt.Sprintf("User (%s) successfully retrieved todo (%d)", userID, ID)
	s.logger.Info(infoMsg)

	return todo, nil, nil
}

func (s *Service) DeleteTodo(todoID int, userID string) (clientError, internalError) {
	// TODO run auth check
	todo, err := s.repo.GetTodoByID(todoID)
	if err != nil {
		errMsg := fmt.Sprintf("Could not get todo (%d) for user (%s)", todoID, userID)
		s.logger.Error(errMsg)
		return nil, err
	}

	if todo == nil {
		clientError := NewClientError("The todo you tried to delete does not exist", http.StatusBadRequest)
		return clientError, nil
	}

	if todo.UserID != userID {
		clientError := NewClientError("You do not have permission to delete this todo", http.StatusUnauthorized)
		return clientError, nil
	}

	err = s.repo.DeleteTodo(todoID)
	if err != nil {
		errMsg := fmt.Sprintf("Could not delete todo (%d)", todoID)
		s.logger.Error(errMsg)
		return nil, err
	}

	infoMsg := fmt.Sprintf("User (%s) succesfully deleted  todo (%d)", userID, todoID)
	s.logger.Info(infoMsg)

	return nil, nil
}

func (s *Service) DeleteAllTodosByUserID(userID string) internalError {
	err := s.repo.DeleteAllTodosByUserID(userID)
	if err != nil {
		errMsg := fmt.Sprintf("Could not delete all todos for user (%s)", userID)
		s.logger.Error(errMsg)
		return err
	}

	infoMsg := fmt.Sprintf("User (%s) successfully deleted all their todos", userID)
	s.logger.Info(infoMsg)

	return nil
}

func (s *Service) DeleteAllTodosByUserIDAndStatus(userID string, IsComplete bool) internalError {
	err := s.repo.DeleteAllTodosByUserIDAndStatus(userID, IsComplete)
	if err != nil {
		errMsg := fmt.Sprintf("Could not delete all todos for user (%s) where completed = %v", userID, IsComplete)
		s.logger.Error(errMsg)
		return err
	}

	infoMsg := fmt.Sprintf("User (%s) deleted all their todos with status %v", userID, IsComplete)
	s.logger.Info(infoMsg)

	return nil
}

func (s *Service) UpdateTodoStatus(userID string, todoID int) (*models.Todo, clientError, internalError) {
	todo, internalErr := s.repo.GetTodoByID(todoID)
	if internalErr != nil {
		errMsg := fmt.Sprintf("Could not get todo (%d)", todoID)
		s.logger.Error(errMsg)
		return nil, nil, internalErr
	}

	if todo == nil {
		clientError := NewClientError("The todo you are updating does not exist", http.StatusBadRequest)
		return nil, clientError, nil
	}

	userIsAuthor := userID == todo.UserID
	if !userIsAuthor {
		warningMsg := fmt.Sprintf("User (%s) attempted to make unauthorized update to todo (%d)", userID, todoID)
		s.logger.Warning(warningMsg)
		clientError := NewClientError("You are not authorized to update this todo", http.StatusUnauthorized)
		return nil, clientError, nil
	}

	updatedStatus := !todo.IsComplete

	todo.IsComplete = updatedStatus

	internalErr = s.repo.UpdateTodo(*todo)
	if internalErr != nil {
		errMsg := fmt.Sprintf("User (%s) could update not todo (%d)", userID, todoID)
		s.logger.Error(errMsg)
		return nil, nil, internalErr
	}

	return todo, nil, nil
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

	infoMsg := fmt.Sprintf("User (%s) created successfully", user.ID)
	s.logger.Info(infoMsg)
	return &user, nil, nil
}

func (s *Service) UserCanCreateNewTodo(user *models.User, list []*models.Todo) (bool, internalError) {
	userIsPaidUser, internalErr := s.UserIsPaidUser(user.ID)
	if internalErr != nil {
		errMsg := fmt.Sprintf("Could not determin payment status for user (%s)", user.ID)
		s.logger.Error(errMsg)
		return false, internalErr
	}

	canCreateNewTodo := (!userIsPaidUser && len(list) < DefaultLimit) || userIsPaidUser

	return canCreateNewTodo, nil
}

func (s *Service) AddStripeIDToUser(userID, stripeID string) internalError {
	internalErr := s.repo.AddStripeIDToUser(userID, stripeID)
	if internalErr != nil {
		errMsg := fmt.Sprintf("Could not add stripe customer id to user (%s)", userID)
		s.logger.Error(errMsg)
	}
	infoMsg := fmt.Sprintf("Stripe ID (%s) added to user (%s)", stripeID, userID)
	s.logger.Info(infoMsg)
	return internalErr
}

func (s *Service) GetUserByID(userID string) (*models.User, internalError) {
	user, internalErr := s.repo.GetUserByID(userID)
	if internalErr != nil {
		errMsg := fmt.Sprintf("Coul not get user (%s)", userID)
		s.logger.Error(errMsg)
	}
	return user, internalErr
}

func (s *Service) GetUserByEmail(email string) (*models.User, internalError) {
	user, internalErr := s.repo.GetUserByEmail(email)
	if internalErr != nil {
		errMsg := fmt.Sprintf("Could not get user (%s)", email)
		s.logger.Error(errMsg)
	}
	return user, internalErr
}

func (s *Service) GetUserByStripeID(customerStripeID string) (*models.User, internalError) {
	user, err := s.repo.GetUserByStripeID(customerStripeID)
	if err != nil {
		errMsg := fmt.Sprintf("Could not get customer from db with stripe ID (%s)", customerStripeID)
		s.logger.Error(errMsg)
	}
	return user, err

}

func (s *Service) UpdateUserPaymentStatus(userID string, isPaidUser bool) internalError {
	internalErr := s.repo.UpdateUserPaymentStatus(userID, isPaidUser)
	if internalErr != nil {
		errMsg := fmt.Sprintf("Could not update user (%s) paymennt status to %v", userID, isPaidUser)
		s.logger.Error(errMsg)
		return internalErr
	}

	infoMsg := fmt.Sprintf("Payment status for user (%s) successfully updated to %v", userID, isPaidUser)
	s.logger.Info(infoMsg)
	return nil
}

func (s *Service) UserIsPaidUser(userID string) (bool, error) {
	isPaidUser, internalErr := s.repo.UserIsPaidUser(userID)
	if internalErr != nil {
		errMsg := fmt.Sprintf("Could not determine payment status of user (%s)", userID)
		s.logger.Error(errMsg)
	}
	return isPaidUser, internalErr
}
