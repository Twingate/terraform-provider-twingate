package twingate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseErrors(t *testing.T) {
	t.Run("Test Twingate Resource : Parse Errors", func(t *testing.T) {

		msg0 := "test response 0"
		msg1 := "test response 1"
		f := []string{msg0, msg1}
		responseLocations := []*queryResponseErrorsLocation{}
		responseLocation := &queryResponseErrorsLocation{
			Line:   1,
			Column: 2,
		}
		responseLocations = append(responseLocations, responseLocation)
		responseErrors := []*queryResponseErrors{}
		var responsePath []string
		responseError0 := &queryResponseErrors{
			Message:   msg0,
			Locations: responseLocations,
			Path:      responsePath,
		}
		responseError1 := &queryResponseErrors{
			Message:   msg1,
			Locations: responseLocations,
			Path:      responsePath,
		}
		responseErrors = append(responseErrors, responseError0)
		responseErrors = append(responseErrors, responseError1)

		messages := parseErrors(responseErrors)

		assert.Equal(t, f, messages)
	})
}
