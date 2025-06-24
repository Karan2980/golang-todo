package auth

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"todo/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validateRegistrationInput(req); err != nil {
		return nil, err
	}

	// Check if user already exists with specific checks
	if err := s.checkUserExists(req.Username, req.Email); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}

	// Insert user into database
	result, err := database.DB.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		req.Username, req.Email, string(hashedPassword),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	// Get the created user ID
	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting user ID: %v", err)
	}

	// Fetch the created user
	user, err := s.getUserByID(int(userID))
	if err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.generateJWTToken(user)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %v", err)
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) Login(req *LoginRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validateLoginInput(req); err != nil {
		return nil, err
	}

	// Check if user exists by email first
	user, err := s.getUserByEmail(req.Email)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, fmt.Errorf("email does not exist")
		}
		return nil, fmt.Errorf("error checking user: %v", err)
	}

	// Get user password for verification
	var hashedPassword string
	err = database.DB.QueryRow("SELECT password FROM users WHERE email = ?", req.Email).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email does not exist")
		}
		return nil, fmt.Errorf("error retrieving user data: %v", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("incorrect password")
	}

	// Generate JWT token
	token, err := s.generateJWTToken(user)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %v", err)
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) Logout(tokenString string) error {
	// Parse and validate the token
	claims, err := s.parseJWTToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %v", err)
	}

	// Check if token is already expired/blacklisted
	if s.isTokenBlacklisted(tokenString) {
		return fmt.Errorf("token is already expired")
	}

	// Get user ID from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	// Add token to blacklist
	_, err = database.DB.Exec(
		"INSERT INTO expired_tokens (user_id, token) VALUES (?, ?)",
		int(userID), tokenString,
	)
	if err != nil {
		return fmt.Errorf("error blacklisting token: %v", err)
	}

	return nil
}

func (s *Service) parseJWTToken(tokenString string) (jwt.MapClaims, error) {
	secretKey := "your-secret-key"
	
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check if token is expired
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token is expired")
			}
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *Service) isTokenBlacklisted(tokenString string) bool {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM expired_tokens WHERE token = ?",
		tokenString,
	).Scan(&count)
	
	if err != nil {
		return false
	}
	
	return count > 0
}

func (s *Service) ValidateToken(tokenString string) (*UserInfo, error) {
	// Check if token is blacklisted
	if s.isTokenBlacklisted(tokenString) {
		return nil, fmt.Errorf("token is blacklisted")
	}

	// Parse and validate token
	claims, err := s.parseJWTToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Get user ID from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Get user info
	user, err := s.getUserByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Service) getUserByID(id int) (*UserInfo, error) {
	var user User
	err := database.DB.QueryRow(
		"SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?", id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *Service) getUserByEmail(email string) (*UserInfo, error) {
	var user User
	err := database.DB.QueryRow(
		"SELECT id, username, email, created_at, updated_at FROM users WHERE email = ?", email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *Service) generateJWTToken(user *UserInfo) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":      time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key (in production, use environment variable)
	secretKey := "your-secret-key"
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *Service) validateRegistrationInput(req *RegisterRequest) error {
	// Validate username
	if strings.TrimSpace(req.Username) == "" {
		return fmt.Errorf("username is required")
	}
	if len(req.Username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(req.Username) > 50 {
		return fmt.Errorf("username must be less than 50 characters")
	}

	// Validate email
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if !s.isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if strings.TrimSpace(req.Password) == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	return nil
}

func (s *Service) validateLoginInput(req *LoginRequest) error {
	// Validate email
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if !s.isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if strings.TrimSpace(req.Password) == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}

func (s *Service) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Updated function to check specific conflicts
func (s *Service) checkUserExists(username, email string) error {
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

// Keep the old function for backward compatibility if needed elsewhere
func (s *Service) userExists(username, email string) (bool, error) {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username = ? OR email = ?",
		username, email,
	).Scan(&count)
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}
