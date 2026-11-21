package httpaccountshandler

import (
	"coderchirag/folio-tracker/lib/domain"
)

type SignUpBody struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	PAN   string `json:"pan"`
}

func (b SignUpBody) ToSignUpSvcInput() *domain.UserParams {
	return &domain.UserParams{
		Name:  b.Name,
		Email: b.Email,
		PAN:   b.PAN,
	}
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	PAN   string `json:"pan"`
}

func (u *UserResponse) FromDomain(d *domain.User) *UserResponse {
	u.ID = d.ID.String()
	u.Name = d.Name
	u.Email = d.Email
	u.PAN = d.PAN
	return u
}
