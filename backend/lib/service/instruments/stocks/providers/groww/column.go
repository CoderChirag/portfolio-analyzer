package stocksgroww

import (
	"strings"
)

func normalizeColumnName(col string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(col)), " ", "_")
}
