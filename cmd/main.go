package main

import (
	"log"
	"net/http"

	"todo/internal/auth"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/models"
	"todo/internal/services"

	"github.com/gorilla/mux"
)

func main() {
	if err := database.Initialize(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Create tables
	if err := models.CreateTodosTable(); err != nil {
		log.Fatal("Failed to create todos table:", err)
	}

	if err := models.CreateUsersTable(); err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	// Initialize services
	todoService := services.NewTodoService()
	authService := auth.NewService()

	// Initialize handlers
	todoHandler := handlers.NewTodoHandler(todoService)
	authHandler := auth.NewHandler(authService)

	// Setup routes
	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Todo routes
	api.HandleFunc("/todos", todoHandler.GetTodos).Methods("GET")
	api.HandleFunc("/todos", todoHandler.CreateTodo).Methods("POST")
	api.HandleFunc("/todos/{id}", todoHandler.GetTodo).Methods("GET")
	api.HandleFunc("/todos/{id}", todoHandler.UpdateTodo).Methods("PUT")
	api.HandleFunc("/todos/{id}", todoHandler.DeleteTodo).Methods("DELETE")

	// Auth routes
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
