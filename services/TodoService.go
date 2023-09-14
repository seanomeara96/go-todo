package services

import (
	"fmt"
	"go-todo/models"
)

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

func (s *Service) DeleteAllTodosByUserIDAndStatus(userID string, IsComplete bool) bool {
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
