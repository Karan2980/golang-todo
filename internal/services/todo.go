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

func (s *TodoService) GetAll(userID int) ([]models.Todo, error) {
	rows, err := database.DB.Query(`
		SELECT id, user_id, title, description, completed, order_no, created_at, updated_at 
		FROM todos 
		WHERE user_id = ? 
		ORDER BY order_no ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed, &todo.OrderNo, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (s *TodoService) GetByID(id, userID int) (*models.Todo, error) {
	var todo models.Todo
	err := database.DB.QueryRow(`
		SELECT id, user_id, title, description, completed, order_no, created_at, updated_at 
		FROM todos 
		WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed, &todo.OrderNo, &todo.CreatedAt, &todo.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo not found")
	}
	return &todo, err
}

func (s *TodoService) Create(todo *models.Todo, userID int) (*models.Todo, error) {
	if todo.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Get the next order number for this user
	nextOrderNo, err := s.getNextOrderNo(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting next order number: %v", err)
	}

	result, err := database.DB.Exec(`
		INSERT INTO todos (user_id, title, description, completed, order_no) 
		VALUES (?, ?, ?, ?, ?)`, 
		userID, todo.Title, todo.Description, todo.Completed, nextOrderNo)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return s.GetByID(int(id), userID)
}

func (s *TodoService) Update(id int, todo *models.Todo, userID int) (*models.Todo, error) {
	if todo.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Check if todo exists and belongs to user
	existingTodo, err := s.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	result, err := database.DB.Exec(`
		UPDATE todos 
		SET title = ?, description = ?, completed = ? 
		WHERE id = ? AND user_id = ?`, 
		todo.Title, todo.Description, todo.Completed, id, userID)
	if err != nil {
		return nil, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, fmt.Errorf("todo not found")
	}

	// Return updated todo with preserved order_no
	updatedTodo := &models.Todo{
		ID:          existingTodo.ID,
		UserID:      existingTodo.UserID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		OrderNo:     existingTodo.OrderNo,
		CreatedAt:   existingTodo.CreatedAt,
		UpdatedAt:   existingTodo.UpdatedAt,
	}

	return updatedTodo, nil
}

func (s *TodoService) Delete(id, userID int) error {
	// Get the todo to be deleted to know its order_no
	todoToDelete, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete the todo
	result, err := tx.Exec("DELETE FROM todos WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("todo not found")
	}

	// Update order numbers for remaining todos
	_, err = tx.Exec(`
		UPDATE todos 
		SET order_no = order_no - 1 
		WHERE user_id = ? AND order_no > ?`, 
		userID, todoToDelete.OrderNo)
	if err != nil {
		return fmt.Errorf("error updating order numbers: %v", err)
	}

	return tx.Commit()
}

func (s *TodoService) ReorderTodos(userID int, todoID int, newOrderNo int) error {
	// Get current todo
	currentTodo, err := s.GetByID(todoID, userID)
	if err != nil {
		return err
	}

	// Get max order number for user
	maxOrderNo, err := s.getMaxOrderNo(userID)
	if err != nil {
		return err
	}

	// Validate new order number
	if newOrderNo < 1 || newOrderNo > maxOrderNo {
		return fmt.Errorf("invalid order number: must be between 1 and %d", maxOrderNo)
	}

	if currentTodo.OrderNo == newOrderNo {
		return nil // No change needed
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	if currentTodo.OrderNo < newOrderNo {
		// Moving down: shift todos up
		_, err = tx.Exec(`
			UPDATE todos 
			SET order_no = order_no - 1 
			WHERE user_id = ? AND order_no > ? AND order_no <= ?`,
			userID, currentTodo.OrderNo, newOrderNo)
	} else {
		// Moving up: shift todos down
		_, err = tx.Exec(`
			UPDATE todos 
			SET order_no = order_no + 1 
			WHERE user_id = ? AND order_no >= ? AND order_no < ?`,
			userID, newOrderNo, currentTodo.OrderNo)
	}

	if err != nil {
		return fmt.Errorf("error updating order numbers: %v", err)
	}

	// Update the target todo's order
	_, err = tx.Exec(`
		UPDATE todos 
		SET order_no = ? 
		WHERE id = ? AND user_id = ?`,
		newOrderNo, todoID, userID)
	if err != nil {
		return fmt.Errorf("error updating todo order: %v", err)
	}

	return tx.Commit()
}

func (s *TodoService) getNextOrderNo(userID int) (int, error) {
	var maxOrderNo sql.NullInt64
	err := database.DB.QueryRow(`
		SELECT MAX(order_no) 
		FROM todos 
		WHERE user_id = ?`, userID).Scan(&maxOrderNo)
	
	if err != nil {
		return 0, err
	}
	
	if maxOrderNo.Valid {
		return int(maxOrderNo.Int64) + 1, nil
	}
	
	return 1, nil // First todo for this user
}

func (s *TodoService) getMaxOrderNo(userID int) (int, error) {
	var maxOrderNo sql.NullInt64
	err := database.DB.QueryRow(`
		SELECT MAX(order_no) 
		FROM todos 
		WHERE user_id = ?`, userID).Scan(&maxOrderNo)
	
	if err != nil {
		return 0, err
	}
	
	if maxOrderNo.Valid {
		return int(maxOrderNo.Int64), nil
	}
	
	return 0, nil
}
