package models

import "github.com/google/uuid"

type User struct {
	Name  string
	Email string
}

func NewUser(email string, name string) User {
	return User{
		Name:  name,
		Email: email,
	}
}

type UserRecord struct {
	ID string
	User
	Password string
}

func NewUserRecord(name, email, password string) UserRecord {
	id := uuid.New().String()
	return UserRecord{
		ID:       id,
		User:     NewUser(email, name),
		Password: password,
	}
}
