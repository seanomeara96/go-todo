package repositories

import (
	"database/sql"
	"go-todo/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewuserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) Save(user models.UserRecord) error {
	stmt, err := r.db.Prepare(`INSERT INTO users(id, name,  email, password) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetUser(ID string) (*models.User, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email FROM users WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(ID).Scan(user.ID, user.Name, user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
