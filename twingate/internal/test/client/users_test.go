package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClientUsersReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Users - Ok", func(t *testing.T) {
		expected := []*User{
			{ID: "user-1", FirstName: "First", LastName: "Last", Email: "user-1@gmail.com", Role: "ADMIN"},
			{ID: "user-2", FirstName: "Second", LastName: "Last", Email: "user-2@gmail.com", Role: "DEVOPS"},
			{ID: "user-3", FirstName: "John", LastName: "White", Email: "user-3@gmail.com", Role: "ADMIN"},
		}

		jsonResponse := `{
	  "data": {
		"users": {
		  "pageInfo": {
			"endCursor": "cursor",
			"hasNextPage": true
		  },
		  "edges": [
			{
			  "node": {
				"id": "user-1",
				"firstName": "First",
				"lastName": "Last",
				"email": "user-1@gmail.com",
				"role": "ADMIN"
			  }
			},
			{
			  "node": {
				"id": "user-2",
				"firstName": "Second",
				"lastName": "Last",
				"email": "user-2@gmail.com",
				"role": "DEVOPS"
			  }
			}
		  ]
		}
	  }
	}`

		nextPage := `{
	  "data": {
		"users": {
		  "pageInfo": {
			"hasNextPage": false
		  },
		  "edges": [
			{
			  "node": {
				"id": "user-3",
				"firstName": "John",
				"lastName": "White",
				"email": "user-3@gmail.com",
				"role": "ADMIN"
			  }
			}
		  ]
		}
	  }
	}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.ResponderFromMultipleResponses([]*http.Response{
				httpmock.NewStringResponse(200, jsonResponse),
				httpmock.NewStringResponse(200, nextPage),
			}),
		)

		users, err := client.ReadUsers(context.Background())

		assert.Nil(t, err)
		assert.Equal(t, expected, users)
	})
}

func TestClientUsersReadEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Read Users - Empty Result", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
			"users": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		users, err := client.ReadUsers(context.Background())

		assert.Nil(t, err)
		assert.Nil(t, users)
		assert.Len(t, users, 0)
	})
}

func TestClientUsersReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Users - Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
			"users": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, jsonResponse)
				return resp, errors.New("error_1")
			})

		users, err := client.ReadUsers(context.Background())

		assert.Nil(t, users)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read user with id All: Post "%s": error_1`, client.GraphqlServerURL))
	})
}
