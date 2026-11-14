package api

import (
	"coderchirag/folio-tracker/lib/api/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func registerRoutes(r *chi.Mux, handlers handlerStore) {
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Post("/signup", handlers.Accounts.SignUp)
	r.Get("/users/{id}", handlers.Accounts.GetUserByID)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.JSONError(w, r, response.ErrNotFound, nil)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.JSONError(w, r, response.ErrMethodNotAllowed, nil)
	})
}
