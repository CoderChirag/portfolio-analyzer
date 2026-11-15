package responsewriter

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
)

type writer struct {
	http.ResponseWriter
	r           *http.Request
	wroteHeader bool
	blocked     bool // set once a second response is attempted; drops all later writes
}

func (w *writer) WriteHeader(code int) {
	if w.blocked {
		return
	}
	if w.wroteHeader {
		w.blocked = true
		slog.ErrorContext(
			w.r.Context(), "double response write blocked",
			"request_id", middleware.GetReqID(w.r.Context()),
			"method", w.r.Method,
			"path", w.r.URL.Path,
			"attempted_status", code,
			"stack", string(debug.Stack()),
		)
		return
	}
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *writer) Write(b []byte) (int, error) {
	if w.blocked {
		return 0, http.ErrHandlerTimeout // any error; the body is dropped
	}
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Lets http.ResponseController reach Flush/Hijack/SetDeadline on the real writer.
func (w *writer) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func ResponseWriter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&writer{
			ResponseWriter: w,
			r:              r,
		}, r)
	})
}
