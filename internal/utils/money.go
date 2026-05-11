package utils

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// ParseAmountToMinor converts a decimal amount string (e.g. "12.34") to integer
// minor currency units (e.g. 1234). Accepts at most 2 decimal places. The
// result must be greater than zero.
func ParseAmountToMinor(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("amount is required")
	}
	if strings.HasPrefix(s, "-") {
		return 0, errors.New("amount must be positive")
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, errors.New("invalid amount format")
	}

	major, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || major < 0 {
		return 0, errors.New("invalid amount format")
	}

	minor := int64(0)
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) == 0 || len(frac) > 2 {
			return 0, errors.New("amount may have at most 2 decimal places")
		}
		if len(frac) == 1 {
			frac += "0"
		}
		minor, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, errors.New("invalid amount format")
		}
	}

	// Guard against int64 overflow before multiplying.
	// math.MaxInt64 == 9223372036854775807; MaxInt64/100 == 92233720368547758,
	// MaxInt64%100 == 7.  If major exceeds that quotient, major*100 overflows.
	// If major equals the quotient, the allowed minor ceiling is MaxInt64%100.
	if major > math.MaxInt64/100 || (major == math.MaxInt64/100 && minor > math.MaxInt64%100) {
		return 0, errors.New("amount too large")
	}

	total := major*100 + minor
	if total <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}
	return total, nil
}
