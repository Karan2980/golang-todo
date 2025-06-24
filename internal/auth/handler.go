package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"todo/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	
	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Register user
	authResp, err := h.service.Register(&req)
	if err != nil {
		// Determine appropriate status code based on error
		statusCode := http.StatusBadRequest
		
		// Check for specific conflict errors
		errorMsg := err.Error()
		if errorMsg == "user with this username already exists" || 
		   errorMsg == "user with this email already exists" {
			statusCode = http.StatusConflict
		}
		
		response.Error(w, err.Error(), statusCode)
		return
	}

	// Return success response
	response.Success(w, "User registered successfully", authResp, http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	
	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Login user
	authResp, err := h.service.Login(&req)
	if err != nil {
		// Determine appropriate status code based on error type
		statusCode := http.StatusUnauthorized
		errorMsg := err.Error()
		
		// Handle validation errors with 400 status
		if strings.Contains(errorMsg, "required") || 
		   strings.Contains(errorMsg, "invalid email format") {
			statusCode = http.StatusBadRequest
		}
		
		response.Error(w, err.Error(), statusCode)
		return
	}

	// Return success response
	response.Success(w, "Login successful", authResp, http.StatusOK)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response.Error(w, "Authorization header is required", http.StatusBadRequest)
		return
	}

	// Check if header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		response.Error(w, "Invalid authorization header format. Use 'Bearer <token>'", http.StatusBadRequest)
		return
	}

	// Extract token from header
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if strings.TrimSpace(token) == "" {
		response.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Logout user (blacklist token)
	if err := h.service.Logout(token); err != nil {
		// Determine status code based on error
		statusCode := http.StatusBadRequest
		if strings.Contains(err.Error(), "invalid token") || 
		   strings.Contains(err.Error(), "token is expired") ||
		   strings.Contains(err.Error(), "token is already expired") {
			statusCode = http.StatusUnauthorized
		}
		response.Error(w, err.Error(), statusCode)
		return
	}

	// Return success response
	response.Success(w, "Logout successful", nil, http.StatusOK)
}

// Helper function to extract token from request (can be used by other handlers)
func (h *Handler) ExtractTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("token is required")
	}

	return token, nil
}
