package api

import (
	"coderchirag/folio-tracker/lib/api/middlewares/auth"
	"coderchirag/folio-tracker/lib/api/middlewares/responsewriter"
	"coderchirag/folio-tracker/lib/api/response"
	"coderchirag/folio-tracker/lib/apperr"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func registerRoutes(r *chi.Mux, handlers handlerStore) {
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		responsewriter.ResponseWriter,
		middleware.Recoverer,
	)

	// Accounts
	r.Post("/signup", handlers.Accounts.SignUp)
	r.With(auth.Auth).Route("/users", func(r chi.Router) {
		// Get User By Email.
		r.Get("/", handlers.Accounts.GetUser)

		// Get User By ID.
		r.Get("/{id}", handlers.Accounts.GetUserByID)

	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.JSONError(w, r, apperr.ErrNotFound, nil)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.JSONError(w, r, response.ErrMethodNotAllowed, nil)
	})
}
