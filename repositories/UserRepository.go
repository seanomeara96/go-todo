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
	stmt, err := r.db.Prepare(`INSERT INTO users(id, name,  email, password, is_paid_user) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsPaidUser)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByID(ID string) (*models.User, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email, is_paid_user FROM users WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(ID).Scan(&user.ID, &user.Name, &user.Email, &user.IsPaidUser)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) AddStripeIDToUser(userID, stripeID string) error {
	stmt, err := r.db.Prepare(`UPDATE users SET customer_stripe_id = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(stripeID, userID)
	return err
}

func (r *UserRepository) UpdateUserPaymentStatus(userID string, isPaidUser bool) error {
	stmt, err := r.db.Prepare(`UPDATE users SETis_paid_user = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(isPaidUser, userID)
	return err
}

func (r *UserRepository) GetUserRecordByEmail(email string) (*models.UserRecord, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email, password, is_paid_user FROM users WHERE email = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	userRecord := models.UserRecord{}
	err = stmt.QueryRow(email).Scan(
		&userRecord.ID,
		&userRecord.Name,
		&userRecord.Email,
		&userRecord.Password,
		&userRecord.IsPaidUser,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &userRecord, nil
}
