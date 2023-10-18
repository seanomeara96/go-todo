package main

import (
	"fmt"
	"go-todo/internal/db"
	"go-todo/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db, err := db.Connect()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, user_id, description, is_complete FROM todos WHERE user_id NOT IN (SELECT id FROM  users)")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		todo := models.Todo{}
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Description, &todo.IsComplete)
		if err != nil {
			panic(err)
		}
		todos = append(todos, todo)
	}

	fmt.Printf("Found %d todos that can be deleted\n", len(todos))

	stmt, err := db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, todo := range todos {
		_, err := stmt.Exec(todo.ID)
		if err != nil {
			panic(err)
		}
	}

}
