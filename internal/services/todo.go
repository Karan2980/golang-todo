package services

import (
	"database/sql"
	"fmt"

	"todo/internal/database"
	"todo/internal/models"
)

type TodoService struct{}

func NewTodoService() *TodoService {
	return &TodoService{}
}

func (s *TodoService) GetAll() ([]models.Todo, error) {
	rows, err := database.DB.Query("SELECT id, title, description, completed, created_at, updated_at FROM todos ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (s *TodoService) GetByID(id int) (*models.Todo, error) {
	var todo models.Todo
	err := database.DB.QueryRow("SELECT id, title, description, completed, created_at, updated_at FROM todos WHERE id = ?", id).
		Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo not found")
	}
	return &todo, err
}

func (s *TodoService) Create(todo *models.Todo) (*models.Todo, error) {
	if todo.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	result, err := database.DB.Exec("INSERT INTO todos (title, description, completed) VALUES (?, ?, ?)", 
		todo.Title, todo.Description, todo.Completed)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return s.GetByID(int(id))
}

func (s *TodoService) Update(id int, todo *models.Todo) (*models.Todo, error) {
	if todo.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	result, err := database.DB.Exec("UPDATE todos SET title = ?, description = ?, completed = ? WHERE id = ?", 
		todo.Title, todo.Description, todo.Completed, id)
	if err != nil {
		return nil, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, fmt.Errorf("todo not found")
	}

	return s.GetByID(id)
}

func (s *TodoService) Delete(id int) error {
	result, err := database.DB.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("todo not found")
	}

	return nil
}
