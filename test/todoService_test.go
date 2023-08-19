package test

import (
	"database/sql"
	"fmt"
	"go-todo/repositories"
	"go-todo/services"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func TestTodoService(t *testing.T) {
	db, err := sql.Open("sqlite3", "todo_service_test.db")
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

	todoRepo := repositories.NewTodoRepo(db)
	if err = todoRepo.Init(); err != nil {
		t.Error(err)
		return
	}

	todoService := services.NewTodoService(todoRepo)

	userId := uuid.New().String()

	list, err := todoService.GetUserTodoList(userId)
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) > 0 {
		t.Error("expected no rows")
		fmt.Println(list)
		return
	}

	_, err = todoService.Create(userId, "new todo")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = todoService.Create(userId, "another new todo")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = todoService.Create(userId, "another another new todo")

	if err != nil {
		t.Error(err)
		return
	}

	todoList, err := todoService.GetUserTodoList(userId)
	if err != nil {
		t.Error(err)
		return
	}

	for _, item := range todoList {
		fmt.Println(item.Description)
	}

}
