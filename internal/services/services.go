package services

import (
	"go-todo/internal/logger"
	"go-todo/internal/repositories"
	"go-todo/internal/server/cache"
)

const DefaultLimit = 10

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
