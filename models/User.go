package models

type User struct {
	ID    string
	Name  string
	Email string
}

func NewUser(ID string, name string, email string) User {
	return User{
		ID:    ID,
		Name:  name,
		Email: email,
	}
}

type UserRecord struct {
	User
	Password string
}

func NewUserRecord(id, name, email, password string) UserRecord {

	return UserRecord{
		User:     NewUser(id, name, email),
		Password: password,
	}
}
