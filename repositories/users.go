package repositories

import (
	"go-todo/models"
)

func (r *Repository) SaveUser(user models.User) error {
	stmt, err := r.db.Prepare(`INSERT INTO users(id, name, email, password, is_paid_user, customer_stripe_id) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsPaidUser, &user.StripeCustomerID)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}

	return nil
}

func (r *Repository) GetUserByID(ID string) (*models.User, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email, password, is_paid_user, customer_stripe_id FROM users WHERE id = ?`)
	if err != nil {
		r.logger.Debug(err.Error())
		return nil, err
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(ID).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsPaidUser, &user.StripeCustomerID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		r.logger.Debug(err.Error())
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email, password, is_paid_user, customer_stripe_id FROM users WHERE email = ?`)
	if err != nil {
		r.logger.Debug(err.Error())
		return nil, err
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsPaidUser, &user.StripeCustomerID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		r.logger.Debug(err.Error())
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UserEmailExists(email string) (bool, error) {
	stmt, err := r.db.Prepare(`SELECT email FROM users WHERE email = ?`)
	if err != nil {
		r.logger.Debug(err.Error())
		return false, err
	}
	defer stmt.Close()

	var matchingEmail string
	err = stmt.QueryRow(email).Scan(&matchingEmail)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		r.logger.Debug(err.Error())
		return false, err
	}
	return true, nil
}

func (r *Repository) GetUserByStripeID(customerStripeID string) (*models.User, error) {
	qry := `SELECT 
				id, 
				name, 
				email, 
				password, 
				is_paid_user, 
				customer_stripe_id 
			FROM users 
			WHERE customer_stripe_id = ?`

	stmt, err := r.db.Prepare(qry)
	if err != nil {
		r.logger.Debug(err.Error())
		return nil, err
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(customerStripeID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.IsPaidUser,
		&user.StripeCustomerID,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		r.logger.Debug(err.Error())
		return nil, err
	}
	return &user, nil
}

func (r *Repository) AddStripeIDToUser(userID, stripeID string) error {
	stmt, err := r.db.Prepare(`UPDATE users SET customer_stripe_id = ? WHERE id = ?`)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(stripeID, userID)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	return nil
}

func (r *Repository) UpdateUserPaymentStatus(userID string, isPaidUser bool) error {
	stmt, err := r.db.Prepare(`UPDATE users SET is_paid_user = ? WHERE id = ?`)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(isPaidUser, userID)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	return nil
}

func (r *Repository) UserIsPaidUser(userID string) (bool, error) {
	stmt, err := r.db.Prepare(`SELECT is_paid_user FROM users WHERE id = ?`)
	if err != nil {
		r.logger.Debug(err.Error())
		return false, err
	}
	defer stmt.Close()

	var isPaidUser bool
	err = stmt.QueryRow(userID).Scan(&isPaidUser)
	if err != nil {
		r.logger.Debug(err.Error())
		return false, err
	}
	return isPaidUser, nil
}
