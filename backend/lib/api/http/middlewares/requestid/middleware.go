package httprequestidmiddlware

import (
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	RequestIDHeader = "X-Request-Id"
	RequestIDKey    = middleware.RequestIDKey
)

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = domain.NewID().String()
		}
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
