package routes

import (
	"todo/internal/handlers"
	"todo/internal/services"

	"github.com/gorilla/mux"
)

func SetupRouter(todoHandler *handlers.TodoHandler, authHandler *handlers.Handler, authService *services.AuthService) *mux.Router {
	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	SetupTodoRoutes(api, todoHandler, authService)
	SetupAuthRoutes(api, authHandler)
	return router
}
