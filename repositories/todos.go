package repositories

import (
	"database/sql"
	"fmt"
	"go-todo/models"
)

func (r *Repository) CreateTodo(todo *models.Todo) (int, error) {
	stmt, err := r.db.Prepare(`INSERT INTO todos(user_id, description, is_complete) VALUES (?, ?, false)`)
	if err != nil {
		r.logger.Debug(err.Error())
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(todo.UserID, todo.Description)
	if err != nil {
		r.logger.Debug(err.Error())
		return 0, err
	}
	_id, err := res.LastInsertId()
	if err != nil {
		r.logger.Debug(err.Error())
		return 0, fmt.Errorf("could not get last insert id")
	}
	return int(_id), nil
}

func (r *Repository) GetTodoByID(ID int) (*models.Todo, error) {
	stmt, err := r.db.Prepare("SELECT id, user_id, description, is_complete FROM todos WHERE id = ?")
	if err != nil {
		r.logger.Debug(err.Error())
		return nil, err
	}
	defer stmt.Close()

	todo := models.Todo{}
	err = stmt.QueryRow(ID).Scan(&todo.ID, &todo.UserID, &todo.Description, &todo.IsComplete)
	if err != nil {
		r.logger.Debug(err.Error())
		return nil, err
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
		r.logger.Debug(err.Error())
		return nil, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	if limit > 0 {
		rows, err = stmt.Query(userID, limit)
	} else {
		rows, err = stmt.Query(userID)
	}
	if err != nil {
		r.logger.Debug(err.Error())
		return nil, err
	}
	defer rows.Close()

	todoList := []*models.Todo{}
	for rows.Next() {
		todo := models.Todo{}
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Description, &todo.IsComplete)
		if err != nil {
			r.logger.Debug(err.Error())
			return nil, err
		}
		todoList = append(todoList, &todo)
	}

	return todoList, nil
}

func (r *Repository) UpdateTodo(todo models.Todo) error {
	stmt, err := r.db.Prepare(`UPDATE todos SET user_id = ?, description = ?, is_complete = ? WHERE id = ?`)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(todo.UserID, todo.Description, todo.IsComplete, todo.ID)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	return nil
}

func (r *Repository) DeleteTodo(todoID int) error {
	stmt, err := r.db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(todoID)
	return err
}

func (r *Repository) DeleteAllTodosByUserID(userID string) error {
	stmt, err := r.db.Prepare("DELETE FROM todos WHERE user_id = ?")
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	return nil
}

func (r *Repository) DeleteAllTodosByUserIDAndStatus(userID string, IsComplete bool) error {
	stmt, err := r.db.Prepare("DELETE FROM todos WHERE user_id = ? AND is_complete = ?")
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, IsComplete)
	if err != nil {
		r.logger.Debug(err.Error())
		return err
	}
	return nil
}
