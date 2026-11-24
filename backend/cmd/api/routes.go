package api

import (
	httpauthmiddleware "coderchirag/folio-tracker/lib/api/http/middlewares/auth"
	httpresponsewriter "coderchirag/folio-tracker/lib/api/http/middlewares/responsewriter"
	httpresponse "coderchirag/folio-tracker/lib/api/http/response"
	"coderchirag/folio-tracker/lib/apperr"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func registerRoutes(r *chi.Mux, handlers handlerStore) {
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		httpresponsewriter.ResponseWriter,
		middleware.Recoverer,
	)

	// Accounts
	r.Post("/signup", handlers.Accounts.SignUp)
	r.With(httpauthmiddleware.Auth).Route("/users", func(r chi.Router) {
		// Get User By Email.
		r.Get("/", handlers.Accounts.GetUser)

		// Get User By ID.
		r.Get("/{id}", handlers.Accounts.GetUserByID)

	})

	// Uploads
	r.With(httpauthmiddleware.Auth).Route("/import", func(r chi.Router) {
		r.Post("/stocks/transactions", handlers.Uploads.ImportStockTransactions)
		r.Get("/status", handlers.Uploads.ImportStatus)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpresponse.JSONError(w, r, apperr.ErrNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpresponse.JSONError(w, r, httpresponse.ErrMethodNotAllowed)
	})
}
