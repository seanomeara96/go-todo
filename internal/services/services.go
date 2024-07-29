package services

import (
	"go-todo/internal/repositories"
	"go-todo/internal/server/cache"
)

const DefaultLimit = 10

type clientError *ClientError

type Service struct {
	repo   *repositories.Repository
	caches *cache.Caches
}

func NewService(r *repositories.Repository, caches *cache.Caches) *Service {
	return &Service{
		repo:   r,
		caches: caches,
	}
}
