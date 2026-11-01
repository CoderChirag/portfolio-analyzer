package utils

import (
	"net/mail"
	"strings"
)

// isBareEmail accepts only "local@domain": no display name, no comments,
// no angle brackets. net/mail alone would accept "Name <a@b.com>".
func IsBareEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s || addr.Name != "" {
		return false
	}
	at := strings.LastIndexByte(s, '@')
	return at > 0 && at < len(s)-1 && strings.Contains(s[at+1:], ".")
}
