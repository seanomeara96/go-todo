package repositories

import (
	"fmt"
	"go-todo/internal/models"
)

const sqlNoResult = "sql: no rows in result set"

func (r *Repository) SaveUser(user models.User) error {
	stmt, err := r.db.Prepare(`INSERT INTO users(id, name, email, password, is_paid_user, customer_stripe_id) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("Issue while preparing save user statement. %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsPaidUser, &user.StripeCustomerID)
	if err != nil {
		return fmt.Errorf("Error while executing save user statement. %w", err)
	}

	return nil
}

func (r *Repository) GetUserByID(ID string) (*models.User, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email, password, is_paid_user, customer_stripe_id FROM users WHERE id = ?`)
	if err != nil {
		return nil, fmt.Errorf("Issue while preparing get user by id statement. %w", err)
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(ID).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsPaidUser, &user.StripeCustomerID)
	if err != nil {
		// TODO perhaps better to include a count() to query and if 0 return nil?
		if err.Error() == sqlNoResult {
			return nil, nil
		}
		return nil, fmt.Errorf("Error shile executing get user by id query. %w", err)
	}

	return &user, nil
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	stmt, err := r.db.Prepare(`SELECT id, name, email, password, is_paid_user, customer_stripe_id FROM users WHERE email = ?`)
	if err != nil {
		return nil, fmt.Errorf("Issue preparing get user by email query. %w", err)
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsPaidUser, &user.StripeCustomerID)
	if err != nil {
		if err.Error() == sqlNoResult {
			return nil, nil
		}
		return nil, fmt.Errorf("Error while executing get user by email query. %w", err)
	}
	return &user, nil
}

func (r *Repository) UserEmailExists(email string) (bool, error) {
	stmt, err := r.db.Prepare(`SELECT email FROM users WHERE email = ?`)
	if err != nil {
		return false, fmt.Errorf("Issue preparing user email exists query. %w", err)
	}
	defer stmt.Close()

	var matchingEmail string
	err = stmt.QueryRow(email).Scan(&matchingEmail)
	if err != nil {
		if err.Error() == sqlNoResult {
			return false, nil
		}
		return false, fmt.Errorf("Issue executing user meail exists query. %w", err)
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
		return nil, fmt.Errorf("Issue preparing get user by stripe id statement. %w", err)
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
		if err.Error() == sqlNoResult {
			return nil, nil
		}
		return nil, fmt.Errorf("Error executing get user by stripe id statement. %w", err)
	}
	return &user, nil
}

func (r *Repository) AddStripeIDToUser(userID, stripeID string) error {
	stmt, err := r.db.Prepare(`UPDATE users SET customer_stripe_id = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("Issue preparing add stripe id to user statement. %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(stripeID, userID)
	if err != nil {
		return fmt.Errorf("Error executing add stripe id to user statement. %w", err)
	}
	return nil
}

func (r *Repository) UpdateUserPaymentStatus(userID string, isPaidUser bool) error {
	stmt, err := r.db.Prepare(`UPDATE users SET is_paid_user = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("Issue preparing update user payment status query. %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(isPaidUser, userID)
	if err != nil {
		return fmt.Errorf("Issue executing update user payment status query. %w", err)
	}
	return nil
}

func (r *Repository) UserIsPaidUser(userID string) (bool, error) {
	stmt, err := r.db.Prepare(`SELECT is_paid_user FROM users WHERE id = ?`)
	if err != nil {
		return false, fmt.Errorf("Issue preparing is paid user query. %w", err)
	}
	defer stmt.Close()

	var isPaidUser bool
	err = stmt.QueryRow(userID).Scan(&isPaidUser)
	if err != nil {
		return false, fmt.Errorf("Error while executing is paid user query. %w", err)
	}
	return isPaidUser, nil
}
