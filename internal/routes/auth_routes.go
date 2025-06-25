package routes

import (
	"todo/internal/handlers"

	"github.com/gorilla/mux"
)

func SetupAuthRoutes(api *mux.Router, authHandler *handlers.Handler) {
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")
}
