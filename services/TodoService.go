package services

import (
	"fmt"
	"go-todo/models"
	repos "go-todo/repositories"
)

type TodoService struct {
	repo *repos.TodoRepo
}

func NewTodoService(repo *repos.TodoRepo) *TodoService {
	return &TodoService{
		repo: repo,
	}
}

func (s *TodoService) Create(userID, description string) (*models.Todo, error) {
	todo := models.NewTodo(userID, description)
	id, err := s.repo.Create(&todo)
	if err != nil {
		return nil, err
	}
	todo.ID = id
	return &todo, nil
}

func (s *TodoService) GetUserTodoList(userID string) ([]*models.Todo, error) {
	return s.repo.GetAll(userID)
}

func (s *TodoService) GetByID(ID int) (*models.Todo, error) {
	return s.repo.Get(ID)
}

func (s *TodoService) Remove(ID int) error {
	// TODO run auth check

	return s.repo.Delete(ID)
}

func (s *TodoService) UpdateStatus(userID string, todoID int) (*models.Todo, error) {
	todo, err := s.repo.Get(todoID)
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

	err = s.repo.Update(*todo)
	if err != nil {
		return nil, err
	}

	return todo, nil

}
