package models

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

type User struct {
	ID        ID        `db:"id"          json:"id"`
	Name      string    `db:"name"        json:"name"`
	Email     string    `db:"email"       json:"email"`
	PAN       string    `db:"pan"         json:"pan"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt time.Time `db:"updated_at"  json:"updated_at"`
}

var panPattern = regexp.MustCompile("^[A-Z]{5}[0-9]{4}[A-Z]$")

// NewUser normalizes and validates input and assigns a fresh ID
func NewUser(name, email, pan string) (*User, error) {
	user := &User{
		ID:    NewID(),
		Name:  strings.TrimSpace(name),
		Email: NormalizeEmail(email),
		PAN:   strings.ToUpper(strings.TrimSpace(pan)),
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) Validate() error {
	if u.Name == "" || len(u.Name) > 100 {
		return errors.New("name must be 1-100 characters")
	}
	if len(u.Email) > 255 {
		return errors.New("email too long")
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return errors.New("invalid email")
	}
	if !panPattern.MatchString(u.PAN) {
		return errors.New("invalid PAN")
	}
	return nil
}

// NormalizeEmail is used both when saving and when searching by email
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
