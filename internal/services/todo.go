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
		s.logger.Error("Could not create new Todo item")
		return nil, nil, err
	}

	todo.ID = lastInsertedTodoID

	infoMsg := fmt.Sprintf("User (%s) successfully created new todo (%d)", userID, todo.ID)
	s.logger.Info(infoMsg)

	return &todo, nil, nil
}

func (s *Service) GetUserTodoList(userID string) ([]*models.Todo, error) {
	user := s.caches.UserCache.GetUserByID(userID)
	if user != nil {
		s.logger.Info("got user by user cache")
	}

	if user == nil {
		u, err := s.repo.GetUserByID(userID)
		if err != nil {
			errMsg := fmt.Sprintf("Something went wrong look for user by ID (%s)", user.ID)
			s.logger.Error(errMsg)
			return nil, err
		}
		user = u
	}

	if user == nil {
		errMsg := fmt.Sprintf("Could not find user (%s)", userID)
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

func (s *Service) GetTodoByID(ID int, userID string) (*models.Todo, clientError, error) {
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

func (s *Service) DeleteTodo(todoID int, userID string) (clientError, error) {
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

func (s *Service) DeleteAllTodosByUserID(userID string) error {
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

func (s *Service) DeleteAllTodosByUserIDAndStatus(userID string, IsComplete bool) error {
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

func (s *Service) UpdateTodoStatus(userID string, todoID int) (*models.Todo, clientError, error) {
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

func (s *Service) DeleteUnattributedTodos() error {
	return s.repo.DeleteUnattributedTodos()
}
