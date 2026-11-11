package utils

import (
	"mime"
	"path"
	"strings"
)

// baseFileName drops any directory part a client sent ("C:\Users\me\orders.csv" → "orders.csv").
func BaseFileName(name string) string {
	name = strings.TrimSpace(strings.ReplaceAll(name, `\`, "/"))
	if name == "" {
		return ""
	}
	base := path.Base(name)
	if base == "." || base == "/" {
		return ""
	}
	return base
}

// normalizeMediaType keeps "type/subtype" in lowercase and drops parameters such as charset.
func NormalizeMediaType(ct string) string {
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(ct))
	}
	return mt
}
