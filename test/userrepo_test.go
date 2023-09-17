package test

import (
	"database/sql"
	"fmt"
	"go-todo/repositories"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestEmailExists(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	repo := repositories.NewRepository(db)

	found, err := repo.UserEmailExists("email@email.com")
	if err != nil {
		t.Error(err)
	}

	if found != true {
		t.Error("should have been found")
	}

	found, err = repo.UserEmailExists("sean@email.com")
	if err != nil {
		fmt.Println(err.Error())
		t.Error(err)
	}

	if found != false {
		t.Error("should have been false")
	}
}

func TestGetUserByStripeID(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	repo := repositories.NewRepository(db)

	user, err := repo.GetUserByStripeID("cus_OctA19z8oZnMu5")
	if err != nil {
		t.Error(err)
	}

	if user == nil {
		t.Error("there should be an associated user")
	}
}
