package models

type User struct {
	ID         string
	Name       string
	Email      string
	IsPaidUser bool
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
