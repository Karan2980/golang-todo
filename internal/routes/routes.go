package routes

import (
	"todo/internal/auth"
	"todo/internal/handlers"

	"github.com/gorilla/mux"
)

func SetupRouter(todoHandler *handlers.TodoHandler, authHandler *handlers.AuthHandler, authService *auth.AuthService) *mux.Router {
	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	SetupTodoRoutes(api, todoHandler, authService)
	SetupAuthRoutes(api, authHandler)
	return router
}
