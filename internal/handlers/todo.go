package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"todo/internal/models"
	"todo/internal/services"
	"todo/pkg/response"

	"github.com/gorilla/mux"
)

type TodoHandler struct {
	service *services.TodoService
}

func NewTodoHandler(service *services.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.service.GetAll()
	if err != nil {
		response.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	response.Success(w, "Todos fetched successfully", todos, http.StatusOK)
}

func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	 // 1. Extract ID from URL path parameter
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}
// 2. Call service layer to get todo by ID
	todo, err := h.service.GetByID(id)
	if err != nil {
		if err.Error() == "todo not found" {
			response.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			response.Error(w, "Failed to fetch todo", http.StatusInternalServerError)
		}
		return
	}
	// 3. Return successful response with todo data
	response.Success(w, "Todo fetched successfully", todo, http.StatusOK)
}

func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		response.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	createdTodo, err := h.service.Create(&todo)
	if err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(w, "Todo created successfully", createdTodo, http.StatusCreated)
}

func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
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

	updatedTodo, err := h.service.Update(id, &todo)
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
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "todo not found" {
			response.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			response.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		}
		return
	}
	response.Success(w, "Todo deleted successfully", nil, http.StatusOK)
}
