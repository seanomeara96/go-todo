package models

type Todo struct {
	ID          int
	UserID      string
	Description string
	IsComplete  bool
}

func NewTodo(userID string, description string) Todo {
	return Todo{
		UserID:      userID,
		Description: description,
	}
}
