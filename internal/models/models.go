package models

type Todo struct {
	ID          int
	UserID      string
	Description string
	IsComplete  bool
}

func NewTodo(userID string, description string) Todo {
	return Todo{
		UserID:      userID,
		Description: description,
	}
}

type User struct {
	ID               string
	Name             string
	Email            string
	Password         string
	IsPaidUser       bool
	StripeCustomerID string
	Roles            map[string]string
}

func NewUser(ID string, name string, email string, password string, isPaidUser bool, stripeCustomerID string) User {
	return User{
		ID:               ID,
		Name:             name,
		Email:            email,
		Password:         password,
		IsPaidUser:       isPaidUser,
		StripeCustomerID: stripeCustomerID,
	}
}
