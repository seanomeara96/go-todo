package models

type User struct {
	ID          string
	Name        string
	Email       string
	IsPayedUser bool
}

func NewUser(ID string, name string, email string, isPayedUser bool) User {
	return User{
		ID:          ID,
		Name:        name,
		Email:       email,
		IsPayedUser: isPayedUser,
	}
}

type UserRecord struct {
	User
	Password string
}

func NewUserRecord(id, name, email, password string, isPayedUser bool) UserRecord {
	return UserRecord{
		User:     NewUser(id, name, email, isPayedUser),
		Password: password,
	}
}
