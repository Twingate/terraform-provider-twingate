package transport

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func Test_idToString(t *testing.T) {
	cases := []struct {
		id       graphql.ID
		expected string
	}{
		{
			id:       nil,
			expected: "",
		},
		{
			id:       graphql.ID("123"),
			expected: "123",
		},
		{
			id:       graphql.ID(101),
			expected: "101",
		},
		{
			id:       graphql.ID(101.5),
			expected: "101.5",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("test case #%d", i+1), func(t *testing.T) {
			actual := idToString(c.id)
			assert.Equal(t, c.expected, actual)
		})
	}
}
