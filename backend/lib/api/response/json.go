package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxbytes int64) error {
	var e *http.MaxBytesError
	r.Body = http.MaxBytesReader(w, r.Body, maxbytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil && errors.As(err, &e) {
		return fmt.Errorf("%w: %v", ErrEntityTooLarge, err)
	} else if err != nil {
		return fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	if dec.More() {
		return fmt.Errorf("%w: multiple JSON values", ErrBadRequest)
	}
	return nil
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": data})
}

func JSONError(w http.ResponseWriter, r *http.Request, err error) {
	status, body := classify(err)
	if status >= 500 {
		slog.ErrorContext(r.Context(), "request failed", "method", r.Method,
			"path", r.URL.Path, "err", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
