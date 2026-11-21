package httpauthmiddleware

import (
	httpresponse "coderchirag/folio-tracker/lib/api/http/response"
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
			httpresponse.JSONError(w, r, ErrMissingXAuthToken, classifyErrors)
			return
		}
		ctx = context.WithValue(ctx, UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
