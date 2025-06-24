package auth

import (
	"encoding/json"
	"net/http"

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
		if err.Error() == "user with this username or email already exists" {
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
		response.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Return success response
	response.Success(w, "Login successful", authResp, http.StatusOK)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// For JWT tokens, logout is typically handled client-side by removing the token
	// But you can implement token blacklisting here if needed
	response.Success(w, "Logout successful", nil, http.StatusOK)
}
