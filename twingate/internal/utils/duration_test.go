package utils

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDurationSupportsDays(t *testing.T) {
	cases := []struct {
		duration         string
		expectedDuration time.Duration
		expectedErr      error
	}{
		{
			duration:         "1d4h5m",
			expectedDuration: 28*time.Hour + 5*time.Minute,
		},
		{
			duration:         "2d3h",
			expectedDuration: 51 * time.Hour,
		},
		{
			duration:         "1.5d",
			expectedDuration: 36 * time.Hour,
		},
		{
			duration:         ".5d",
			expectedDuration: 12 * time.Hour,
		},
		{
			duration:    "1.5d1d",
			expectedErr: errors.New("invalid day duration: 1.5d1d"),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%d", i+1), func(t *testing.T) {
			duration, err := ParseDurationWithDays(c.duration)
			if c.expectedErr != nil {
				assert.EqualError(t, err, c.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expectedDuration, duration)
			}
		})
	}
}

func TestFormatDurationWithDays(t *testing.T) {
	cases := []struct {
		duration time.Duration
		expected string
	}{
		{
			duration: 51*time.Hour + 5*time.Minute + 30*time.Second,
			expected: "2d3h5m30s",
		},
		{
			duration: 51 * time.Hour,
			expected: "2d3h",
		},
		{
			duration: 36 * time.Hour,
			expected: "1d12h",
		},
		{
			duration: 12 * time.Hour,
			expected: "12h",
		},
		{
			duration: 0,
			expected: "0s",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%d", i+1), func(t *testing.T) {
			actual := FormatDurationWithDays(c.duration)
			assert.Equal(t, c.expected, actual)
		})
	}
}
