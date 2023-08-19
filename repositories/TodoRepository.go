package repositories

import (
	"database/sql"
	"go-todo/models"
)

type TodoRepo struct {
	db *sql.DB
}

func NewTodoRepo(db *sql.DB) *TodoRepo {
	return &TodoRepo{
		db: db,
	}
}

func (r *TodoRepo) Create(todo *models.Todo) error {
	stmt, err := r.db.Prepare(`INSERT INTO todos(user_id, description, is_complete) VALUES (?, ?, false)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(todo.UserID, todo.Description)
	if err != nil {
		return err
	}
	return nil
}

func (r *TodoRepo) Get(ID int) (*models.Todo, error) {
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

func (r *TodoRepo) GetAll(userID string) ([]*models.Todo, error) {
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

func (r *TodoRepo) Update(todo models.Todo) error {
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

func (r *TodoRepo) Delete(todoID int) error {
	stmt, err := r.db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(todoID)
	return err
}

func (r *TodoRepo) Init() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS todos(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		description TEXT DEFAULT "",
		is_complete BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`)
	return err
}
