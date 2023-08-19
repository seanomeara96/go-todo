package services

import (
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
	err := s.repo.Create(&todo)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (s *TodoService) GetUserTodoList(userID string) ([]*models.Todo, error) {
	return s.repo.GetAll(userID)
}

func (s *TodoService) GetByID(ID int) (*models.Todo, error) {
	return s.repo.Get(ID)
}

func (s *TodoService) Remove(ID int) error {
	return s.repo.Delete(ID)
}
