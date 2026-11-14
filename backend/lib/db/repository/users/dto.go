package users

import (
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/domain"
)

func ToUserModel(d *domain.User) *model.Users {
	u := &model.Users{
		ID:        d.ID,
		Name:      d.Name,
		Email:     d.Email,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		Version:   d.Version,
	}

	if d.PAN != "" {
		u.Pan = &d.PAN
	}

	return u
}

func ToUserDomain(m *model.Users) *domain.User {
	u := &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Version:   m.Version,
	}

	if m.Pan != nil {
		u.PAN = *m.Pan
	}

	return u
}
