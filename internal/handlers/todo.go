package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"todo/internal/auth"
	"todo/internal/models"
	"todo/internal/services"
	"todo/pkg/response"

	"github.com/gorilla/mux"
)

type TodoHandler struct {
	service     *services.TodoService
	authService *auth.Service
}

func NewTodoHandler(service *services.TodoService, authService *auth.Service) *TodoHandler {
	return &TodoHandler{
		service:     service,
		authService: authService,
	}
}

// Helper function to extract and validate user from token
func (h *TodoHandler) getUserFromToken(r *http.Request) (*auth.UserInfo, error) {
	// Get token from Authorization header
	token := r.Header.Get("Authorization")
	
	// If token starts with "Bearer ", remove it for backward compatibility
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("authorization token is required")
	}

	// Validate token and get user info
	user, err := h.authService.ValidateToken(strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	user, err := h.getUserFromToken(r)
	if err != nil {
		response.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	todos, err := h.service.GetAll(user.ID)
	if err != nil {
		response.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	response.Success(w, "Todos fetched successfully", todos, http.StatusOK)
}

func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	user, err := h.getUserFromToken(r)
	if err != nil {
		response.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path parameter
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	// Call service layer to get todo by ID for specific user
	todo, err := h.service.GetByID(id, user.ID)
	if err != nil {
		if err.Error() == "todo not found" {
			response.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			response.Error(w, "Failed to fetch todo", http.StatusInternalServerError)
		}
		return
	}

	// Return successful response with todo data
	response.Success(w, "Todo fetched successfully", todo, http.StatusOK)
}

func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	user, err := h.getUserFromToken(r)
	if err != nil {
		response.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		response.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	createdTodo, err := h.service.Create(&todo, user.ID)
	if err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(w, "Todo created successfully", createdTodo, http.StatusCreated)
}

func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	user, err := h.getUserFromToken(r)
	if err != nil {
		response.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		response.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updatedTodo, err := h.service.Update(id, &todo, user.ID)
	if err != nil {
		if err.Error() == "todo not found" {
			response.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			response.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	response.Success(w, "Todo updated successfully", updatedTodo, http.StatusOK)
}

func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	user, err := h.getUserFromToken(r)
	if err != nil {
		response.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id, user.ID); err != nil {
		if err.Error() == "todo not found" {
			response.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			response.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		}
		return
	}
	response.Success(w, "Todo deleted successfully", nil, http.StatusOK)
}

// New handler for reordering todos
func (h *TodoHandler) ReorderTodo(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	user, err := h.getUserFromToken(r)
	if err != nil {
		response.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	var reorderRequest struct {
		NewOrderNo int `json:"new_order_no"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reorderRequest); err != nil {
		response.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.ReorderTodos(user.ID, id, reorderRequest.NewOrderNo); err != nil {
		if err.Error() == "todo not found" {
			response.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			response.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	response.Success(w, "Todo reordered successfully", nil, http.StatusOK)
}
