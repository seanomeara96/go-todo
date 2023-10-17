package repositories

import (
	"database/sql"
	"go-todo/internal/logger"
)

type Repository struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewRepository(db *sql.DB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}
