package query

import (
	"fmt"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

func Test_idToString(t *testing.T) {
	cases := []struct {
		id       graphql.ID
		expected string
	}{
		{
			id:       graphql.ID("123"),
			expected: "123",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("test case #%d", i+1), func(t *testing.T) {
			actual := idToString(c.id)
			assert.Equal(t, c.expected, actual)
		})
	}
}
