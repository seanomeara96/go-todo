package test

import (
	"database/sql"
	"fmt"
	"go-todo/models"
	"go-todo/repositories"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func TestGetTodos(t *testing.T) {
	db, err := sql.Open("sqlite3", "todo_repo_test.db")
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("DROP TABLE IF EXISTS todos")
	if err != nil {
		t.Error(err)
		return
	}

	ID := uuid.New().String()
	name := "example-name"
	email := "example@example.com"
	user := models.NewUser(ID, name, email)

	todoRepo := repositories.NewTodoRepo(db)
	err = todoRepo.Init()
	if err != nil {
		t.Error(err)
		return
	}

	todo := models.NewTodo(user.ID, "my new todo")
	err = todoRepo.Create(&todo)
	if err != nil {
		t.Error(err)
		return
	}

	todo = models.NewTodo(user.ID, "another todo")
	err = todoRepo.Create(&todo)
	if err != nil {
		t.Error(err)
		return
	}

	todoList, err := todoRepo.GetAll(user.ID)
	if err != nil {
		t.Error(err)
		return
	}

	if len(todoList) != 2 {
		t.Error("expected 2 items")
	}
	fmt.Println(todoList[0].UserID)
	t.Error()
}
