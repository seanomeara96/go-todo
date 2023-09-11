package services

import (
	"fmt"
	"go-todo/models"
)

func (s *Service) Create(userID, description string) (*models.Todo, error) {
	todo := models.NewTodo(userID, description)
	id, err := s.repo.CreateTodo(&todo)
	if err != nil {
		return nil, err
	}
	todo.ID = id
	return &todo, nil
}

func (s *Service) GetUserTodoList(userID string) ([]*models.Todo, error) {
	return s.repo.GetAllTodosByUserID(userID)
}

func (s *Service) GetByID(ID int) (*models.Todo, error) {
	return s.repo.GetTodoByID(ID)
}

func (s *Service) Remove(ID int) error {
	// TODO run auth check

	return s.repo.DeleteTodo(ID)
}

func (s *Service) UpdateStatus(userID string, todoID int) (*models.Todo, error) {
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
