package main

import (
	"log"
	"net/http"

	"todo/internal/auth"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/models"
	"todo/internal/routes"
	"todo/internal/services"
)

func main() {
	if err := database.Initialize(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Create tables in correct order (users first, then todos, then expired_tokens)
	if err := auth.CreateUsersTable(); err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	if err := models.CreateTodosTable(); err != nil {
		log.Fatal("Failed to create todos table:", err)
	}

	if err := auth.CreateExpiredTokensTable(); err != nil {
		log.Fatal("Failed to create expired_tokens table:", err)
	}

	// Initialize services
	todoService := services.NewTodoService()
	authService := auth.NewService()

	// Initialize handlers (todoHandler)
	todoHandler := handlers.NewTodoHandler(todoService)   
	authHandler := auth.NewHandler(authService)

	router := routes.SetupRouter(todoHandler, authHandler, authService)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
