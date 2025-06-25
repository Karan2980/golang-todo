package routes

import (
	"net/http"
	"todo/internal/auth"
	"todo/internal/handlers"
	"todo/internal/middleware"

	"github.com/gorilla/mux"
)

func SetupTodoRoutes(api *mux.Router, todoHandler *handlers.TodoHandler, authService *auth.Service) {
	api.Handle("/todos", middleware.AuthMiddleware(authService)(http.HandlerFunc(todoHandler.GetTodos))).Methods("GET")
	api.Handle("/todos", middleware.AuthMiddleware(authService)(http.HandlerFunc(todoHandler.CreateTodo))).Methods("POST")
	api.Handle("/todos/{id}", middleware.AuthMiddleware(authService)(http.HandlerFunc(todoHandler.GetTodo))).Methods("GET")
	api.Handle("/todos/{id}", middleware.AuthMiddleware(authService)(http.HandlerFunc(todoHandler.UpdateTodo))).Methods("PUT")
	api.Handle("/todos/{id}", middleware.AuthMiddleware(authService)(http.HandlerFunc(todoHandler.DeleteTodo))).Methods("DELETE")
	api.Handle("/todos/{id}/reorder", middleware.AuthMiddleware(authService)(http.HandlerFunc(todoHandler.ReorderTodo))).Methods("PUT")
	// api.HandleFunc("/todos", todoHandler.CreateTodo).Methods("POST")
	// api.HandleFunc("/todos/{id}", todoHandler.GetTodo).Methods("GET")
	// api.HandleFunc("/todos/{id}", todoHandler.UpdateTodo).Methods("PUT")
	// api.HandleFunc("/todos/{id}", todoHandler.DeleteTodo).Methods("DELETE")
	// api.HandleFunc("/todos/{id}/reorder", todoHandler.ReorderTodo).Methods("PUT")
} 