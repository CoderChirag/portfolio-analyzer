package httpauthmiddleware

import (
	httpresponse "coderchirag/folio-tracker/lib/api/http/response"
	"coderchirag/folio-tracker/lib/domain"
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
		id, err := domain.ParseID(userID)
		if err != nil {
			httpresponse.JSONError(w, r, httpresponse.ErrUnauthorized, classifyErrors)
			return
		}
		ctx = context.WithValue(ctx, UserIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
