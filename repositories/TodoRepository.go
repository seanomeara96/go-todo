package repositories

import (
	"fmt"
	"go-todo/models"
)

func (r *Repository) CreateTodo(todo *models.Todo) (int, error) {
	stmt, err := r.db.Prepare(`INSERT INTO todos(user_id, description, is_complete) VALUES (?, ?, false)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(todo.UserID, todo.Description)
	if err != nil {
		return 0, err
	}
	_id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not get last insert id")
	}
	return int(_id), nil
}

func (r *Repository) GetTodoByID(ID int) (*models.Todo, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, description, is_complete FROM todos WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	todo := models.Todo{}
	err = stmt.QueryRow(ID).Scan(&todo.ID, &todo.UserID, &todo.Description, &todo.IsComplete)
	if err != nil {
		return nil, err
	}
	return &todo, nil

}

func (r *Repository) GetAllTodosByUserID(userID string) ([]*models.Todo, error) {
	stmt, err := r.db.Prepare(`SELECT id, user_id, description, is_complete FROM todos WHERE user_id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todoList := []*models.Todo{}
	for rows.Next() {
		todo := models.Todo{}
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Description, &todo.IsComplete)
		if err != nil {
			return nil, err
		}
		todoList = append(todoList, &todo)
	}

	return todoList, nil
}

func (r *Repository) UpdateTodo(todo models.Todo) error {
	stmt, err := r.db.Prepare(`UPDATE todos SET user_id = ?, description = ?, is_complete = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(todo.UserID, todo.Description, todo.IsComplete, todo.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteTodo(todoID int) error {
	stmt, err := r.db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(todoID)
	return err
}
