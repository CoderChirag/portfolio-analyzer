package utils

import "time"

// dateOnly keeps t's calendar date (in t's own location) as midnight UTC,
// which is how a DATE column scans with loc=UTC.
func DateOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func ParseTimeWithIST(layout string, timestamp string) (time.Time, error) {
	ist := time.FixedZone("IST", 5*3600+30*60)
	return time.ParseInLocation(layout, timestamp, ist)
}
