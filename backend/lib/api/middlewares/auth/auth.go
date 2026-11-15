package auth

import (
	"coderchirag/folio-tracker/lib/api/response"
	"context"
	"net/http"
)

const (
	UserIDKey = "userID"
)

func Auth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := r.Header.Get("X-Auth-Token")
		if userID == "" {
			response.JSONError(w, r, ErrMissingXAuthToken, classifyErrors)
			return
		}
		ctx = context.WithValue(ctx, UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
