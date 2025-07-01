package database

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

	_, err := DB.Exec(query)
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

	_, err := DB.Exec(query)
	return err
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

	_, err := DB.Exec(query)
	return err
} 