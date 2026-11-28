package utils

import (
	"math/rand/v2"
	"time"
)

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

// jitter returns a random duration in [d/2, d]
func Jitter(d time.Duration) time.Duration {
	if d <= 1 {
		return d
	}
	half := d / 2
	return half + rand.N(half+1)
}

// backoff gives 30s, 1m, 2m, 4m, … capped at 1h, with jitter. attempt is at least 1 after a claim.
func Backoff(minRetry, maxRetry time.Duration, attempt uint16) time.Duration {
	shift := min(max(attempt, 1)-1, 16)
	return Jitter(min(minRetry<<shift, maxRetry))
}
