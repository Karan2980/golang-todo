package models

import (
	"time"
	"todo/internal/database"
)

type Todo struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	OrderNo     int       `json:"order_no"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func CreateTodosTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		order_no INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		INDEX idx_user_id (user_id),
		INDEX idx_user_order (user_id, order_no),
		UNIQUE KEY unique_user_order (user_id, order_no)
	)`

	_, err := database.DB.Exec(query)
	return err
}
