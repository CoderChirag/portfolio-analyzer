package utils

import "strings"

// tokenize splits on anything that isn't an ASCII letter, digit or '&',
// and also splits letter/digit boundaries so "Nifty50" == "Nifty 50".
func Tokenize(s string) []string {
	var tokens []string
	var curr strings.Builder
	var prev rune

	flush := func() {
		if curr.Len() > 0 {
			tokens = append(tokens, curr.String())
			curr.Reset()
		}
	}

	for _, r := range s {
		switch {
		case r == '&':
			flush()
			tokens = append(tokens, "&")
		case IsAsciiDigit(r) || IsAsciiLetter(r):
			if curr.Len() > 0 && IsAsciiLetter(prev) != IsAsciiLetter(r) {
				flush()
			}
			curr.WriteRune(r)
		default:
			flush() // spaces, punctuation, and any non-ASCII left after folding
		}
		prev = r
	}
	flush()
	return tokens
}

// truncateKey caps the key at the column width without cutting a token in half.
func TruncateKey(key string, maxLen int) string {
	if len(key) <= maxLen {
		return key
	}
	key = key[:maxLen]
	if i := strings.LastIndexByte(key, ' '); i > 0 {
		key = key[:i]
	}
	return key
}

func IsAsciiLetter(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func IsAsciiDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// TruncateRunes caps s at n characters without splitting a multi-byte character.
func TruncateRunes(s string, n int) string {
	count := 0
	for i := range s {
		if count == n {
			return s[:i]
		}
		count++
	}
	return s
}
