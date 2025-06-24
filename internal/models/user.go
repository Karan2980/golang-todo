package models

import (
	"time"
	"todo/internal/database"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Don't include password in JSON responses
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateUsersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`

	_, err := database.DB.Exec(query)
	return err
}
