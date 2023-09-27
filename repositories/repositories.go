package repositories

import (
	"database/sql"
	"go-todo/logger"

	"github.com/patrickmn/go-cache"
)

type Repository struct {
	db     *sql.DB
	cache  *cache.Cache
	logger *logger.Logger
}

func NewRepository(db *sql.DB, cache *cache.Cache, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}
