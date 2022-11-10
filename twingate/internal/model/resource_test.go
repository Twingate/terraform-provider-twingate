package model

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPortRange(t *testing.T) {
	invalidPortsRange := func(str ...string) error {
		port, input := str[0], str[0]
		if len(str) > 1 {
			port = str[1]
		}

		return fmt.Errorf("failed to parse protocols port range \"%s\": port `%s` is not a valid integer: strconv.ParseInt: parsing \"%s\": invalid syntax", input, port, port)
	}

	cases := []struct {
		input       string
		expected    *PortRange
		expectedErr error
	}{
		{
			input:    "80",
			expected: &PortRange{Start: 80, End: 80},
		},
		{
			input:    "80-90",
			expected: &PortRange{Start: 80, End: 90},
		},
		{
			input:       "",
			expectedErr: invalidPortsRange(""),
		},
		{
			input:       " ",
			expectedErr: invalidPortsRange(" "),
		},
		{
			input:       "foo",
			expectedErr: invalidPortsRange("foo"),
		},
		{
			input:       "80-",
			expectedErr: invalidPortsRange("80-", ""),
		},
		{
			input:       "-80",
			expectedErr: invalidPortsRange("-80", ""),
		},
		{
			input:       "80-90-100",
			expectedErr: errors.New("failed to parse protocols port range \"80-90-100\": port range expects 2 values"),
		},
		{
			input:       "80-70",
			expectedErr: errors.New("failed to parse protocols port range \"80-70\": ports 80, 70 needs to be in a rising sequence"),
		},
		{
			input:       "0-65536",
			expectedErr: errors.New("failed to parse protocols port range \"0-65536\": port 65536 not in the range of 0-65535"),
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual, err := NewPortRange(c.input)

			assert.Equal(t, c.expected, actual)

			if c.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, c.expectedErr.Error())
			}
		})
	}
}
