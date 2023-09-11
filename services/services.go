package services

import "go-todo/repositories"

type Service struct {
	repo *repositories.Repository
}

func NewService(r *repositories.Repository) *Service {
	return &Service{
		repo: r,
	}
}
