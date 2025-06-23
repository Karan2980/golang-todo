package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database connection
	err := InitializeDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer CloseDatabase()

	// Create todos table if it doesn't exist
	err = CreateTodosTable()
	if err != nil {
		log.Fatal("Failed to create todos table: ", err)
	}

	log.Println("Database connected and todos table ready!")

	// Setup routes
	router := mux.NewRouter()
	setupTodoRoutes(router)

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func setupTodoRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Todo routes
	api.HandleFunc("/todos", GetTodos).Methods("GET")
	api.HandleFunc("/todos", CreateTodo).Methods("POST")
	api.HandleFunc("/todos/{id}", GetTodo).Methods("GET")
	api.HandleFunc("/todos/{id}", UpdateTodo).Methods("PUT")
	api.HandleFunc("/todos/{id}", DeleteTodo).Methods("DELETE")
}
