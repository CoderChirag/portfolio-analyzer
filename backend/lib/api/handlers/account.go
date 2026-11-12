package handlers

import (
	"coderchirag/folio-tracker/lib/api/response"
	"coderchirag/folio-tracker/lib/service/accounts"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type AccountsHandler struct {
	svc AccountService
}

func NewAccountsHandler(svc AccountService) *AccountsHandler {
	return &AccountsHandler{
		svc: svc,
	}
}

func (h *AccountsHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var body SignUpBody
	if err := response.DecodeJSON(w, r, &body, maxBodyBytes); err != nil {
		response.JSONError(w, r, err, nil)
		return
	}

	u, err := h.svc.SignUp(r.Context(), accounts.SignUpInput{
		Name:  body.Name,
		Email: body.Email,
		PAN:   body.PAN,
	})
	if err != nil {
		response.JSONError(w, r, err, classifyErrors)
		return
	}

	w.Header().Set("Location", "/users/"+u.ID.String())
	response.JSON(w, http.StatusCreated, &SignupResponse{
		ID:    u.ID.String(),
		Name:  u.Name,
		Email: u.Email,
	})
}

func (h *AccountsHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	u, err := h.svc.GetUserByID(r.Context(), id)
	if err != nil {
		response.JSONError(w, r, err, classifyErrors)
		return
	}

	response.JSON(w, http.StatusOK, u)
}
