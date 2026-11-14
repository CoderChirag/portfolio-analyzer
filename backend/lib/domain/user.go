package domain

import (
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/utils"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Limits mirror the users table columns.
const (
	userNameMaxLen  = 255 // VARCHAR(255), counted in characters
	userEmailMaxLen = 254 // RFC 5321 practical maximum; column is VARCHAR(255)
)

var (
	ErrUserNameRequired  = errors.New("name is required")
	ErrUserNameTooLong   = errors.New("name must be at most 255 characters")
	ErrUserNameInvalid   = errors.New("name contains invalid characters")
	ErrUserEmailRequired = errors.New("email is required")
	ErrUserEmailTooLong  = errors.New("email must be at most 254 characters")
	ErrUserEmailInvalid  = errors.New("email is not a valid address")
	ErrUserPANInvalid    = errors.New("PAN must look like ABCDE1234F")
	ErrUserEmailTaken    = errors.New("email is already registered")
	ErrUserPANTaken      = errors.New("PAN is already registered")
)

type User struct {
	ID        ID
	Name      string
	Email     string
	PAN       string
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   uint32
}

type UserParams struct {
	Name  string
	Email string
	PAN   string
}

var panPattern = regexp.MustCompile("^[A-Z]{5}[0-9]{4}[A-Z]$")

// NewUser normalizes and validates input and assigns a fresh ID.
// On failure it returns apperr.ValidationError listing every invalid field.
func NewUser(p UserParams) (*User, error) {
	user := &User{
		ID:    NewID(),
		Name:  strings.TrimSpace(p.Name),
		Email: NormalizeEmail(p.Email),
		PAN:   strings.ToUpper(strings.TrimSpace(p.PAN)),
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}

// Validate checks every field and reports all problems, not just the first.
// Exported so updates (not just creation) can re-validate.
func (u *User) Validate() error {
	var errs apperr.ValidationError

	switch n := utf8.RuneCountInString(u.Name); {
	case n == 0:
		errs.Add("name", ErrUserNameRequired)
	case n > userNameMaxLen:
		errs.Add("name", ErrUserNameTooLong)
	case strings.IndexFunc(u.Name, unicode.IsControl) >= 0:
		errs.Add("name", ErrUserNameInvalid)
	}

	switch {
	case u.Email == "":
		errs.Add("email", ErrUserEmailRequired)
	case len(u.Email) > userEmailMaxLen:
		errs.Add("email", ErrUserEmailTooLong)
	case !utils.IsBareEmail(u.Email):
		errs.Add("email", ErrUserEmailInvalid)
	}

	if u.PAN != "" && !panPattern.MatchString(u.PAN) {
		errs.Add("pan", ErrUserPANInvalid)
	}

	return errs.Err()
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
