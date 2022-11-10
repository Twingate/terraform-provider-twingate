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

func TestClientUserReadOk(t *testing.T) {
	testData := []struct {
		role    string
		isAdmin bool
	}{
		{role: "ADMIN", isAdmin: true},
		{role: "DEVOPS", isAdmin: false},
	}

	for _, td := range testData {
		t.Run("Test Twingate Resource : Read User Ok - "+td.role, func(t *testing.T) {
			const (
				userID = "id"
				email  = "user@email"
			)
			jsonResponse := fmt.Sprintf(`{
			  "data": {
			    "user": {
			      "id": "%s",
			      "firstName": "First",
			      "lastName": "Last",
			      "email": "%s",
			      "role": "%s"
			    }
			  }
			}`, userID, email, td.role)

			client := newHTTPMockClient()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", client.GraphqlServerURL,
				httpmock.NewStringResponder(200, jsonResponse))

			user, err := client.readUser(context.Background(), userID)

			assert.Nil(t, err)
			assert.NotNil(t, user)
			assert.EqualValues(t, userID, user.ID)
			assert.EqualValues(t, email, user.Email)
			assert.EqualValues(t, td.role, user.Role)
			assert.EqualValues(t, td.isAdmin, user.IsAdmin())
		})
	}
}

func TestClientUserReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read User Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "user": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const userID = "userID"
		user, err := client.readUser(context.Background(), userID)

		assert.Nil(t, user)
		assert.EqualError(t, err, fmt.Sprintf("failed to read user with id %s: query result is empty", userID))
	})
}

func TestClientUserReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read User Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "user": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, jsonResponse)
				return resp, errors.New("error_1")
			})
		const userID = "userID"

		user, err := client.readUser(context.Background(), userID)

		assert.Nil(t, user)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read user with id %s: Post "%s": error_1`, userID, client.GraphqlServerURL))
	})
}

func TestClientReadEmptyUserError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty User Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "user": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		user, err := client.readUser(context.Background(), "")

		assert.EqualError(t, err, "failed to read user: id is empty")
		assert.Nil(t, user)
	})
}
