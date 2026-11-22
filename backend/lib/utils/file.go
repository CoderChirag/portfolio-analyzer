package utils

import (
	"bytes"
	"coderchirag/folio-tracker/lib/apperr"
	"io"
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

// readLimited streams r into a single buffer,
// sized up front so it doesn't keep growing and copying.
func ReadLimited(r io.Reader, limit, sizeHint int64) ([]byte, error) {
	var buf bytes.Buffer
	if sizeHint > 0 {
		// The +1 matches the LimitReader below, and MinRead covers the free space
		// ReadFrom requires before each read; without it, the final read would double the buffer.
		buf.Grow(int(min(sizeHint, limit+1)) + bytes.MinRead)
	}
	n, err := buf.ReadFrom(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if n > limit {
		return nil, apperr.ErrFileTooLarge
	}
	return buf.Bytes(), nil
}
