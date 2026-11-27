package httpauthmiddleware

import (
	httpresponse "coderchirag/folio-tracker/lib/api/http/response"
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"net/http"
)

const (
	UserIDHeader = "X-Auth-Token"
	UserIDKey    = "userID"
)

func Auth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := r.Header.Get(UserIDHeader)
		if userID == "" {
			httpresponse.JSONError(w, r, httpresponse.ErrMissingXAuthToken)
			return
		}
		id, err := domain.ParseID(userID)
		if err != nil {
			httpresponse.JSONError(w, r, httpresponse.ErrUnauthorized)
			return
		}
		ctx = context.WithValue(ctx, UserIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
