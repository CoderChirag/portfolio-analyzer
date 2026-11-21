package accountshandler

import (
	"coderchirag/folio-tracker/lib/api/middlewares/auth"
	"coderchirag/folio-tracker/lib/api/response"
	"coderchirag/folio-tracker/lib/apperr"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type Handler struct {
	svc AccountService
}

func New(svc AccountService) *Handler {
	return &Handler{
		svc: svc,
	}
}

// Creates a new user.
func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var body SignUpBody
	if err := response.DecodeJSON(w, r, &body, maxBodyBytes); err != nil {
		response.JSONError(w, r, err, nil)
		return
	}

	u, err := h.svc.SignUp(r.Context(), body.ToSignUpSvcInput())
	if err != nil {
		response.JSONError(w, r, err, classifyErrors)
		return
	}

	response.JSON(w, http.StatusCreated, (&User{}).FromDomain(u))
}

// Gets user by the id provided in URL Params.
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := r.Context().Value(auth.UserIDKey)
	if userID != id {
		response.JSONError(w, r, response.ErrForbidden, classifyErrors)
		return
	}

	u, err := h.svc.GetUserByID(r.Context(), id)
	if err != nil {
		response.JSONError(w, r, err, classifyErrors)
		return
	}

	response.JSON(w, http.StatusOK, (&User{}).FromDomain(u))
}

// Gets User. Needs email in query params.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	email := query.Get("email")
	if email == "" {
		var errs apperr.ValidationError
		errs.Add("email", ErrEmailRequired)
		response.JSONError(w, r, errs, classifyErrors)
		return
	}

	u, err := h.svc.GetUserByEmail(r.Context(), email)
	if err != nil {
		response.JSONError(w, r, err, classifyErrors)
		return
	}

	userID := r.Context().Value(auth.UserIDKey)
	if userID != u.ID.String() {
		response.JSONError(w, r, response.ErrForbidden, classifyErrors)
		return
	}

	response.JSON(w, http.StatusOK, (&User{}).FromDomain(u))
}
