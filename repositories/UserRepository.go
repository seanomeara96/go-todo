package repositories

import "database/sql"

type UserRepository struct {
	db *sql.DB
}

func NewuserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}
