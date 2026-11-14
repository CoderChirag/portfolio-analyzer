package accountshandler

import (
	"coderchirag/folio-tracker/lib/domain"
	"coderchirag/folio-tracker/lib/service/accounts"
)

type SignUpBody struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	PAN   string `json:"pan"`
}

func (b SignUpBody) ToSignUpSvcInput() *accounts.SignUpInput {
	return &accounts.SignUpInput{
		Name:  b.Name,
		Email: b.Email,
		PAN:   b.PAN,
	}
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	PAN   string `json:"pan"`
}

func (u User) FromDomain(d *domain.User) User {
	u.ID = d.ID.String()
	u.Name = d.Name
	u.Email = d.Email
	u.PAN = d.PAN
	return u
}
