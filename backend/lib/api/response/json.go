package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
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

// Wrapper over r.ParseMultipartForm.
// Parses multipart form (files > maxFileSizeInMemoryBytes spill to temp disk automatically).
// Writes 400 json error response if parsing error occur
func ParseMultipartForm(w http.ResponseWriter, r *http.Request, maxMemory int64) error {
	// Parse multipart form (files > maxFileSizeInMemoryBytes spill to temp disk automatically)
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		JSONError(w, r, ErrMultipartParse, Classify)
		return err
	}
	return nil
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": data})
}

func JSONError(w http.ResponseWriter, r *http.Request, err error, classifyFunc func(error) (int, *ErrorBody)) {
	var status int
	var body *ErrorBody
	if classifyFunc != nil {
		status, body = classifyFunc(err)
	} else {
		status, body = Classify(err)
	}

	if status >= 500 || slog.Default().Enabled(r.Context(), slog.LevelDebug) {
		slog.ErrorContext(r.Context(), "request failed",
			"request_id", middleware.GetReqID(r.Context()),
			"method", r.Method, "path", r.URL.Path, "err", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
