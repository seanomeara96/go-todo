package services

import repos "go-todo/repositories"

type TodoService struct {
	repo *repos.TodoRepo
}

func NewTodoService(repo *repos.TodoRepo) *TodoService {
	return &TodoService{
		repo: repo,
	}
}
