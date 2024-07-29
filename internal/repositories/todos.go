package repositories

import (
	"database/sql"
	"fmt"
	"go-todo/internal/models"
)

func (r *Repository) CreateTodo(todo *models.Todo) (int, error) {
	stmt, err := r.db.Prepare(`INSERT INTO todos(user_id, description, is_complete) VALUES (?, ?, false)`)
	if err != nil {
		return 0, fmt.Errorf("Issue preparing insert statement for create todo. %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(todo.UserID, todo.Description)
	if err != nil {
		return 0, fmt.Errorf("Issu executing insert statement for create todo. %w", err)
	}
	_id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not get last insert id. %w", err)
	}
	return int(_id), nil
}

func (r *Repository) GetTodoByID(ID int) (*models.Todo, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, description, is_complete FROM todos WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("Issue preparing statment for getting todo by id. %w", err)
	}
	defer stmt.Close()

	todo := models.Todo{}
	err = stmt.QueryRow(ID).Scan(&todo.ID, &todo.UserID, &todo.Description, &todo.IsComplete)
	if err != nil {
		return nil, fmt.Errorf("Issue executing statement for get todo by id. %w", err)
	}
	return &todo, nil
}

func (r *Repository) GetTodosByUserID(userID string, limit int) ([]*models.Todo, error) {
	query := `SELECT id, user_id, description, is_complete FROM todos WHERE user_id = ?`
	if limit > 0 {
		query += ` limit ?`
	}
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("Issue preparing statement for getting todos by user id. %w", err)
	}
	defer stmt.Close()

	// TODO refactor this. I know how to do this better
	var rows *sql.Rows
	if limit > 0 {
		rows, err = stmt.Query(userID, limit)
	} else {
		rows, err = stmt.Query(userID)
	}
	if err != nil {
		return nil, fmt.Errorf("Error while querying todos by user id. %w", err)
	}
	defer rows.Close()

	todoList := []*models.Todo{}
	for rows.Next() {
		todo := models.Todo{}
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Description, &todo.IsComplete)
		if err != nil {
			return nil, fmt.Errorf("Issue scanning todos. %w", err)
		}
		todoList = append(todoList, &todo)
	}

	return todoList, nil
}

func (r *Repository) UpdateTodo(todo models.Todo) error {
	stmt, err := r.db.Prepare(`UPDATE todos SET user_id = ?, description = ?, is_complete = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("Issue preparing statement for updating todos. %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(todo.UserID, todo.Description, todo.IsComplete, todo.ID)
	if err != nil {
		return fmt.Errorf("Error while executing update todo statement. %w", err)
	}
	return nil
}

func (r *Repository) DeleteTodo(todoID int) error {
	stmt, err := r.db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		return fmt.Errorf("Issue while preparing statement to delete todo. %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(todoID)
	if err != nil {
		return fmt.Errorf("Error executing delete todo statement. %w", err)
	}
	return nil
}

func (r *Repository) DeleteAllTodosByUserID(userID string) error {
	stmt, err := r.db.Prepare("DELETE FROM todos WHERE user_id = ?")
	if err != nil {
		return fmt.Errorf("Issue preparing statement to delete todos by user id. %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	if err != nil {
		return fmt.Errorf("Error executing statement to delete todos by user id. %w", err)
	}
	return nil
}

func (r *Repository) DeleteAllTodosByUserIDAndStatus(userID string, IsComplete bool) error {
	stmt, err := r.db.Prepare("DELETE FROM todos WHERE user_id = ? AND is_complete = ?")
	if err != nil {
		return fmt.Errorf("Issue preparing statement for deleteing user's todos by status. %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, IsComplete)
	if err != nil {
		return fmt.Errorf("Error executing statement to delete user todos by status. %w", err)
	}
	return nil
}

func (r *Repository) DeleteUnattributedTodos() error {
	_, err := r.db.Exec("DELETE FROM todos WHERE user_id NOT IN (SELECT id FROM  users)")
	if err != nil {
		return fmt.Errorf("Error deleting todos where user does not exist. %w", err)
	}
	return nil
}
