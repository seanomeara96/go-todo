package services

import (
	"fmt"
	"go-todo/internal/models"
	"html"
	"net/http"
)

type createTodoClientErrors struct {
	DescriptionErrors []string
}

func (s *Service) CreateTodo(userID, description string) (*models.Todo, *createTodoClientErrors, error) {
	clientErrors := createTodoClientErrors{}
	if description == "" {
		clientErrors.DescriptionErrors = append(clientErrors.DescriptionErrors, "cannot supply an empty description")
	}

	if len(clientErrors.DescriptionErrors) > 0 {
		return nil, &clientErrors, nil
	}

	sanitizedDescription := html.EscapeString(description)

	todo := models.NewTodo(userID, sanitizedDescription)

	lastInsertedTodoID, err := s.repo.CreateTodo(&todo)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not create new todo. %w", err)
	}

	todo.ID = lastInsertedTodoID

	return &todo, nil, nil
}

func (s *Service) GetUserTodoList(userID string) ([]*models.Todo, error) {
	user := s.caches.UserCache.GetUserByID(userID)

	if user == nil {
		u, err := s.repo.GetUserByID(userID)
		if err != nil {
			return nil, fmt.Errorf("Could not get user by ID. %w", err)
		}
		user = u
	}

	// Redundant?
	if user == nil {
		return nil, fmt.Errorf("Could not get user by id. %w")
	}

	limit := DefaultLimit
	if user.IsPaidUser {
		limit = 0
	}

	todoList, err := s.repo.GetTodosByUserID(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("Could not get user by ID. %w", err)
	}

	return todoList, nil
}

func (s *Service) GetTodoByID(ID int, userID string) (*models.Todo, clientError, error) {
	todo, err := s.repo.GetTodoByID(ID)
	// internal server error
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get todo by ID. %w", err)
	}

	// client error
	if todo.UserID != userID {
		return nil, NewClientError("User not authorized", http.StatusUnauthorized), nil
	}

	return todo, nil, nil
}

func (s *Service) DeleteTodo(todoID int, userID string) (clientError, error) {
	// TODO run auth check
	todo, err := s.repo.GetTodoByID(todoID)
	if err != nil {
		return nil, fmt.Errorf("Could not get todo by ID. %w", err)
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
		return nil, fmt.Errorf("Could not delete todo. %w", err)
	}

	return nil, nil
}

func (s *Service) DeleteAllTodosByUserID(userID string) error {
	err := s.repo.DeleteAllTodosByUserID(userID)
	if err != nil {
		return fmt.Errorf("Could not delete all todos by user ID. %w", err)
	}

	return nil
}

func (s *Service) DeleteAllTodosByUserIDAndStatus(userID string, IsComplete bool) error {
	err := s.repo.DeleteAllTodosByUserIDAndStatus(userID, IsComplete)
	if err != nil {
		return fmt.Errorf("Could not delete all user todos by status. %w", err)
	}

	return nil
}

func (s *Service) UpdateTodoStatus(userID string, todoID int) (*models.Todo, clientError, error) {
	todo, err := s.repo.GetTodoByID(todoID)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get todo by ID. %w", err)
	}

	if todo == nil {
		clientError := NewClientError("The todo you are updating does not exist", http.StatusBadRequest)
		return nil, clientError, nil
	}

	userIsAuthor := userID == todo.UserID
	if !userIsAuthor {
		clientError := NewClientError("You are not authorized to update this todo", http.StatusUnauthorized)
		return nil, clientError, nil
	}

	updatedStatus := !todo.IsComplete

	todo.IsComplete = updatedStatus

	err = s.repo.UpdateTodo(*todo)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not update todo status. %w", err)
	}

	return todo, nil, nil
}

func (s *Service) DeleteUnattributedTodos() error {
	return s.repo.DeleteUnattributedTodos()
}
