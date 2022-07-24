package twingate

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
		jsonResponse := `{
	  "data": {
		"users": {
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

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		users, err := client.readUsers(context.Background())

		assert.Nil(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 2)

		assert.EqualValues(t, "user-1", users[0].ID)
		assert.EqualValues(t, "First", users[0].FirstName)
		assert.EqualValues(t, "Last", users[0].LastName)
		assert.EqualValues(t, "user-1@gmail.com", users[0].Email)
		assert.EqualValues(t, "ADMIN", users[0].Role)
		assert.EqualValues(t, true, users[0].IsAdmin())

		assert.EqualValues(t, "user-2", users[1].ID)
		assert.EqualValues(t, "Second", users[1].FirstName)
		assert.EqualValues(t, "Last", users[1].LastName)
		assert.EqualValues(t, "user-2@gmail.com", users[1].Email)
		assert.EqualValues(t, "DEVOPS", users[1].Role)
		assert.EqualValues(t, false, users[1].IsAdmin())
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

		users, err := client.readUsers(context.Background())

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

		users, err := client.readUsers(context.Background())

		assert.Nil(t, users)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read user with id All: Post "%s": error_1`, client.GraphqlServerURL))
	})
}
