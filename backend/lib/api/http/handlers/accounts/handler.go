package httpaccountshandler

import (
	httpauthmiddleware "coderchirag/folio-tracker/lib/api/http/middlewares/auth"
	httpresponse "coderchirag/folio-tracker/lib/api/http/response"
	"coderchirag/folio-tracker/lib/apperr"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type HTTPHandler struct {
	svc AccountService
}

func New(svc AccountService) *HTTPHandler {
	return &HTTPHandler{
		svc: svc,
	}
}

// Creates a new user.
func (h *HTTPHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var body SignUpBody
	if err := httpresponse.DecodeJSON(w, r, &body, maxBodyBytes); err != nil {
		httpresponse.JSONError(w, r, err, nil)
		return
	}

	u, err := h.svc.SignUp(r.Context(), body.ToSignUpSvcInput())
	if err != nil {
		httpresponse.JSONError(w, r, err, classifyErrors)
		return
	}

	httpresponse.JSON(w, http.StatusCreated, (&User{}).FromDomain(u))
}

// Gets user by the id provided in URL Params.
func (h *HTTPHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := r.Context().Value(httpauthmiddleware.UserIDKey)
	if userID != id {
		httpresponse.JSONError(w, r, httpresponse.ErrForbidden, classifyErrors)
		return
	}

	u, err := h.svc.GetUserByID(r.Context(), id)
	if err != nil {
		httpresponse.JSONError(w, r, err, classifyErrors)
		return
	}

	httpresponse.JSON(w, http.StatusOK, (&User{}).FromDomain(u))
}

// Gets User. Needs email in query params.
func (h *HTTPHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	email := query.Get("email")
	if email == "" {
		var errs apperr.ValidationError
		errs.Add("email", ErrEmailRequired)
		httpresponse.JSONError(w, r, errs, classifyErrors)
		return
	}

	u, err := h.svc.GetUserByEmail(r.Context(), email)
	if err != nil {
		httpresponse.JSONError(w, r, err, classifyErrors)
		return
	}

	userID := r.Context().Value(httpauthmiddleware.UserIDKey)
	if userID != u.ID.String() {
		httpresponse.JSONError(w, r, httpresponse.ErrForbidden, classifyErrors)
		return
	}

	httpresponse.JSON(w, http.StatusOK, (&User{}).FromDomain(u))
}
