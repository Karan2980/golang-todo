package auth

import "todo/internal/database"

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

func CreateExpiredTokensTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS expired_tokens (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		token TEXT NOT NULL,
		expired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`

	_, err := database.DB.Exec(query)
	return err
}
