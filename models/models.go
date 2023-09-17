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
	IsPaidUser       bool
	StripeCustomerID string
}

func NewUser(ID string, name string, email string, isPaidUser bool) User {
	return User{
		ID:         ID,
		Name:       name,
		Email:      email,
		IsPaidUser: isPaidUser,
	}
}

type UserRecord struct {
	User
	Password string
}

func NewUserRecord(id, name, email, password string, isPaidUser bool) UserRecord {
	return UserRecord{
		User:     NewUser(id, name, email, isPaidUser),
		Password: password,
	}
}
