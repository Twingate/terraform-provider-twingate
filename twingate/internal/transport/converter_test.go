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

func Test_convertToGQL(t *testing.T) {
	cases := []struct {
		val      interface{}
		expected interface{}
	}{
		{
			val:      nil,
			expected: nil,
		},
		{
			val:      "123",
			expected: graphql.String("123"),
		},
		{
			val:      101,
			expected: graphql.Int(101),
		},
		{
			val:      int32(102),
			expected: graphql.Int(102),
		},
		{
			val:      int64(103),
			expected: graphql.Int(103),
		},
		{
			val:      101.5,
			expected: graphql.Float(101.5),
		},
		{
			val:      true,
			expected: graphql.Boolean(true),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("test case #%d", i+1), func(t *testing.T) {
			actual := tryConvertToGQL(c.val)
			assert.Equal(t, c.expected, actual)
		})
	}
}
