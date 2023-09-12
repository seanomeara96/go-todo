package repositories

import (
	"go-todo/models"
	"log"
)

func (r *Repository) SaveUser(user models.UserRecord) error {
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

func (r *Repository) GetUserByID(ID string) (*models.User, error) {
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

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email, is_paid_user FROM users WHERE email = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(email).Scan(&user.ID, &user.Name, &user.Email, &user.IsPaidUser)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) AddStripeIDToUser(userID, stripeID string) error {
	stmt, err := r.db.Prepare(`UPDATE users SET customer_stripe_id = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(stripeID, userID)
	return err
}

func (r *Repository) UpdateUserPaymentStatus(userID string, isPaidUser bool) error {
	stmt, err := r.db.Prepare(`UPDATE users SET is_paid_user = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(isPaidUser, userID)
	return err
}

func (r *Repository) UserIsPaidUser(userID string) (bool, error) {
	stmt, err := r.db.Prepare(`SELECT is_paid_user FROM users WHERE id = ?`)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var isPaidUser bool
	err = stmt.QueryRow(userID).Scan(&isPaidUser)
	if err != nil {
		return false, err
	}
	return isPaidUser, nil
}

func (r *Repository) GetUserRecordByEmail(email string) (*models.UserRecord, error) {
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
