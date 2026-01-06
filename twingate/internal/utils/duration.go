package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const hoursInDay = 24
const secondsInHour = 3600

// Matches day components like "1d", "1.5d", ".5d" (case-insensitive).
var dayDurationRegexp = regexp.MustCompile(`(?i)([0-9]*\.?[0-9]+)\s*d`)

// ParseDurationWithDays extends time.ParseDuration with support for day units (e.g. "2d", "1d12h").
func ParseDurationWithDays(value string) (time.Duration, error) {
	matches := dayDurationRegexp.FindAllStringSubmatchIndex(value, -1)
	if len(matches) == 0 {
		return time.ParseDuration(value) //nolint:wrapcheck
	}

	if len(matches) > 1 {
		return 0, fmt.Errorf("invalid day duration: %v", value) //nolint:err113
	}

	match := matches[0]
	daysStr := value[match[2]:match[3]]

	days, err := strconv.ParseFloat(strings.TrimSpace(daysStr), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid day duration %s: %w", daysStr, err)
	}

	var duration time.Duration

	restDuration := value[len(daysStr)+1:]
	if restDuration != "" {
		duration, err = time.ParseDuration(restDuration)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", value, err)
		}
	}

	return duration + time.Second*time.Duration(secondsInHour*hoursInDay*days), nil
}

// FormatDurationWithDays converts a duration to a string using a day component when >= 24h.
// Example: 51h -> "2d3h", 36h -> "1d12h".
func FormatDurationWithDays(value time.Duration) string {
	if value == 0 {
		return "0s"
	}

	totalSeconds := int64(value.Seconds())
	secondsPerDay := int64(secondsInHour * hoursInDay)

	days := totalSeconds / secondsPerDay
	remainder := totalSeconds % secondsPerDay

	var builder strings.Builder

	if days > 0 {
		builder.WriteString(strconv.FormatInt(days, 10))
		builder.WriteString("d")
	}

	if remainder > 0 {
		rest := (time.Duration(remainder) * time.Second).String()
		if strings.HasSuffix(rest, "m0s") {
			rest = strings.TrimSuffix(rest, "0s")
		}

		if strings.HasSuffix(rest, "h0m") {
			rest = strings.TrimSuffix(rest, "0m")
		}

		builder.WriteString(rest)
	}

	return builder.String()
}
