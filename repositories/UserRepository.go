package repositories

import (
	"database/sql"
	"go-todo/models"
	"log"
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

func (r *UserRepository) GetUserByID(ID string) (*models.User, error) {
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

func (r *UserRepository) GetUserRecordByEmail(email string) (*models.UserRecord, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email, password FROM users WHERE email = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	userRecord := models.UserRecord{}
	err = stmt.QueryRow(email).Scan(&userRecord.ID, &userRecord.Name, &userRecord.Email, &userRecord.Password)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &userRecord, nil
}

func (r *UserRepository) Init() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS users(
		id TEXT PRIMARY KEY UNIQUE NOT NULL,
		name TEXT DEFAULT "",
		email TEXT DEFAULT "",
		password TEXT DEFAULT "")`)
	return err
}
