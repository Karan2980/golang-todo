package services

import (
	"database/sql"
	"fmt"
	"todo/internal/database"
	"todo/internal/models"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) GetUserByID(id int) (*models.UserInfo, error) {
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?", id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &models.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *UserService) GetUserByEmail(email string) (*models.UserInfo, error) {
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, created_at, updated_at FROM users WHERE email = ?", email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &models.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *UserService) CheckUserExists(username, email string) error {
	// Check if username exists
	var usernameCount int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username = ?",
		username,
	).Scan(&usernameCount)
	if err != nil {
		return fmt.Errorf("error checking username existence: %v", err)
	}
	if usernameCount > 0 {
		return fmt.Errorf("user with this username already exists")
	}

	// Check if email exists
	var emailCount int
	err = database.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE email = ?",
		email,
	).Scan(&emailCount)
	if err != nil {
		return fmt.Errorf("error checking email existence: %v", err)
	}
	if emailCount > 0 {
		return fmt.Errorf("user with this email already exists")
	}

	return nil
}

// InsertUser inserts a new user into the users table
func InsertUser(username, email, hashedPassword string) (sql.Result, error) {
	return database.DB.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, hashedPassword,
	)
}

// GetUserPasswordByEmail fetches the hashed password for a user by email
func GetUserPasswordByEmail(email string, hashedPassword *string) error {
	return database.DB.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(hashedPassword)
}

// InsertExpiredToken adds a token to the expired_tokens table
func InsertExpiredToken(userID int, token string) (sql.Result, error) {
	return database.DB.Exec(
		"INSERT INTO expired_tokens (user_id, token) VALUES (?, ?)",
		userID, token,
	)
}

// CountExpiredTokens counts the number of expired tokens matching a token string
func CountExpiredTokens(token string, count *int) error {
	return database.DB.QueryRow(
		"SELECT COUNT(*) FROM expired_tokens WHERE token = ?",
		token,
	).Scan(count)
}

// CountUsers counts users by username or email
func CountUsers(username, email string, count *int) error {
	return database.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username = ? OR email = ?",
		username, email,
	).Scan(count)
} 