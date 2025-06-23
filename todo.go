package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Todo represents a todo item
type Todo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTodosTable creates the todos table if it doesn't exist
func CreateTodosTable() error {
	if err := EnsureConnection(); err != nil {
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating todos table: %v", err)
	}

	log.Println("Todos table created/verified successfully")
	return nil
}

// GetTodos handles GET /api/v1/todos
func GetTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	log.Println("Getting all todos...")

	if err := EnsureConnection(); err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, fmt.Sprintf("Database connection error: %v", err), http.StatusInternalServerError)
		return
	}

	query := "SELECT id, title, description, completed, created_at, updated_at FROM todos ORDER BY created_at DESC"
	rows, err := DB.Query(query)
	if err != nil {
		log.Printf("Error querying todos: %v", err)
		http.Error(w, fmt.Sprintf("Error querying todos: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Initialize as empty slice instead of nil slice
	todos := make([]Todo, 0)
	
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning todo: %v", err)
			http.Error(w, fmt.Sprintf("Error scanning todo: %v", err), http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating todos: %v", err)
		http.Error(w, fmt.Sprintf("Error iterating todos: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d todos", len(todos))
	
	// This will now return [] instead of null when empty
	json.NewEncoder(w).Encode(todos)
}

// GetTodo handles GET /api/v1/todos/{id}
func GetTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid todo ID: %s", vars["id"])
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	log.Printf("Getting todo with ID: %d", id)

	if err := EnsureConnection(); err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, fmt.Sprintf("Database connection error: %v", err), http.StatusInternalServerError)
		return
	}

	var todo Todo
	query := "SELECT id, title, description, completed, created_at, updated_at FROM todos WHERE id = ?"
	err = DB.QueryRow(query, id).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Todo not found with ID: %d", id)
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		log.Printf("Error querying todo: %v", err)
		http.Error(w, fmt.Sprintf("Error querying todo: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Found todo: %s", todo.Title)
	json.NewEncoder(w).Encode(todo)
}

// CreateTodo handles POST /api/v1/todos
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	log.Println("Creating new todo...")

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		log.Printf("Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received todo data: Title=%s, Description=%s, Completed=%t", todo.Title, todo.Description, todo.Completed)

	// Validate required fields
	if todo.Title == "" {
		log.Println("Title is required")
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if err := EnsureConnection(); err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, fmt.Sprintf("Database connection error: %v", err), http.StatusInternalServerError)
		return
	}

	query := "INSERT INTO todos (title, description, completed) VALUES (?, ?, ?)"
	result, err := DB.Exec(query, todo.Title, todo.Description, todo.Completed)
	if err != nil {
		log.Printf("Error creating todo: %v", err)
		http.Error(w, fmt.Sprintf("Error creating todo: %v", err), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting inserted ID: %v", err)
		http.Error(w, fmt.Sprintf("Error getting inserted ID: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Todo created with ID: %d", id)

	// Fetch the created todo
	var createdTodo Todo
	query = "SELECT id, title, description, completed, created_at, updated_at FROM todos WHERE id = ?"
	err = DB.QueryRow(query, id).Scan(&createdTodo.ID, &createdTodo.Title, &createdTodo.Description, &createdTodo.Completed, &createdTodo.CreatedAt, &createdTodo.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching created todo: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching created todo: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created todo: %s", createdTodo.Title)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTodo)
}

// UpdateTodo handles PUT /api/v1/todos/{id}
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid todo ID: %s", vars["id"])
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	log.Printf("Updating todo with ID: %d", id)

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		log.Printf("Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if todo.Title == "" {
		log.Println("Title is required")
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if err := EnsureConnection(); err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, fmt.Sprintf("Database connection error: %v", err), http.StatusInternalServerError)
		return
	}

	query := "UPDATE todos SET title = ?, description = ?, completed = ? WHERE id = ?"
	result, err := DB.Exec(query, todo.Title, todo.Description, todo.Completed, id)
	if err != nil {
		log.Printf("Error updating todo: %v", err)
		http.Error(w, fmt.Sprintf("Error updating todo: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error checking affected rows: %v", err)
		http.Error(w, fmt.Sprintf("Error checking affected rows: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("Todo not found with ID: %d", id)
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Fetch the updated todo
	var updatedTodo Todo
	query = "SELECT id, title, description, completed, created_at, updated_at FROM todos WHERE id = ?"
	err = DB.QueryRow(query, id).Scan(&updatedTodo.ID, &updatedTodo.Title, &updatedTodo.Description, &updatedTodo.Completed, &updatedTodo.CreatedAt, &updatedTodo.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching updated todo: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching updated todo: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated todo: %s", updatedTodo.Title)
	json.NewEncoder(w).Encode(updatedTodo)
}

// DeleteTodo handles DELETE /api/v1/todos/{id}
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid todo ID: %s", vars["id"])
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting todo with ID: %d", id)

	if err := EnsureConnection(); err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, fmt.Sprintf("Database connection error: %v", err), http.StatusInternalServerError)
		return
	}

	query := "DELETE FROM todos WHERE id = ?"
	result, err := DB.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting todo: %v", err)
		http.Error(w, fmt.Sprintf("Error deleting todo: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error checking affected rows: %v", err)
		http.Error(w, fmt.Sprintf("Error checking affected rows: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("Todo not found with ID: %d", id)
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	log.Printf("Successfully deleted todo with ID: %d", id)
	w.WriteHeader(http.StatusNoContent)
}
