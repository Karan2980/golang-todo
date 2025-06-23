package main

import (
	"log"
	"net/http"

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

	if err := models.CreateTodosTable(); err != nil {
		log.Fatal("Failed to create todos table:", err)
	}

	todoService := services.NewTodoService()
	todoHandler := handlers.NewTodoHandler(todoService)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	
	api.HandleFunc("/todos", todoHandler.GetTodos).Methods("GET")
	api.HandleFunc("/todos", todoHandler.CreateTodo).Methods("POST")
	api.HandleFunc("/todos/{id}", todoHandler.GetTodo).Methods("GET")
	api.HandleFunc("/todos/{id}", todoHandler.UpdateTodo).Methods("PUT")
	api.HandleFunc("/todos/{id}", todoHandler.DeleteTodo).Methods("DELETE")

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
