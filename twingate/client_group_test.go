package twingate

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func TestClientGroupCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Create Group Ok", func(t *testing.T) {
		// response JSON
		createGroupOkJson := `{
		  "data": {
			"groupCreate": {
			  "entity": {
				"id": "test-id"
			  },
			  "ok": true,
			  "error": null
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createGroupOkJson))
		groupName := graphql.String("test")

		group, err := client.createGroup(context.Background(), groupName)

		assert.Nil(t, err)
		assert.EqualValues(t, "test-id", group.ID)
	})
}

func TestClientGroupCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Group Error", func(t *testing.T) {
		// response JSON
		createGroupOkJson := `{
		  "data": {
			"groupCreate": {
			  "ok": false,
			  "error": "error_1"
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createGroupOkJson))
		groupName := graphql.String("test")

		group, err := client.createGroup(context.Background(), groupName)

		assert.EqualError(t, err, "failed to create group: error_1")
		assert.Nil(t, group)
	})
}

func TestClientGroupCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Group Request Error", func(t *testing.T) {
		// response JSON
		createGroupOkJson := `{
		  "data": {
			"groupCreate": {
			  "ok": false,
			  "error": "error_1"
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, createGroupOkJson)
				return resp, errors.New("error_1")
			})

		groupName := graphql.String("test")

		group, err := client.createGroup(context.Background(), groupName)

		assert.EqualError(t, err, "failed to create group: Post \""+client.GraphqlServerURL+"\": error_1")
		assert.Nil(t, group)
	})
}

func TestClientGroupUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Error", func(t *testing.T) {
		// response JSON
		updateGroupOkJson := `{
		  "data": {
			"groupUpdate": {
			  "ok": false,
			  "error": "error_1"
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, updateGroupOkJson))
		groupName := graphql.String("test")
		groupId := graphql.ID("id")
		err := client.updateGroup(context.Background(), groupId, groupName)

		assert.EqualError(t, err, "failed to update group with id id: error_1")
	})
}

func TestClientGroupUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Request Error", func(t *testing.T) {
		// response JSON
		updateGroupOkJson := `{
		  "data": {
			"groupUpdate": {
			  "ok": false,
			  "error": "error_1"
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, updateGroupOkJson)
				return resp, errors.New("error_1")
			})

		groupName := graphql.String("test")
		groupId := graphql.ID("id")
		err := client.updateGroup(context.Background(), groupId, groupName)

		assert.EqualError(t, err, "failed to update group with id id: Post \""+client.GraphqlServerURL+"\": error_1")
	})
}

func TestClientGroupReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Error", func(t *testing.T) {
		// response JSON
		readGroupOkJson := `{
		  "data": {
			"group": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))
		groupId := graphql.ID("id")

		group, err := client.readGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, "failed to read group with id id")
	})
}

func TestClientGroupReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Request Error", func(t *testing.T) {
		// response JSON
		readGroupOkJson := `{
		  "data": {
			"group": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, readGroupOkJson)
				return resp, errors.New("error_1")
			})
		groupId := graphql.ID("id")

		group, err := client.readGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, "failed to read group with id id: Post \""+client.GraphqlServerURL+"\": error_1")
	})
}

func TestClientCreateEmptyGroupError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Empty Group Error", func(t *testing.T) {
		// response JSON
		readGroupOkJson := `{
		  "data": {
			"group": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))
		groupName := graphql.String("")

		group, err := client.createGroup(context.Background(), groupName)

		assert.EqualError(t, err, "failed to create group: group name is empty")
		assert.Nil(t, group)
	})
}

func TestClientReadEmptyGroupError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty Group Error", func(t *testing.T) {
		// response JSON
		readGroupOkJson := `{
		  "data": {
			"group": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))
		groupId := graphql.ID("")

		group, err := client.readGroup(context.Background(), groupId)

		assert.EqualError(t, err, "failed to read group: group id is empty")
		assert.Nil(t, group)
	})
}

func TestClientDeleteEmptyGroupError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Empty Group Error", func(t *testing.T) {
		// response JSON
		readGroupOkJson := `{
		  "data": {
			"group": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))
		groupId := graphql.ID("")

		err := client.deleteGroup(context.Background(), groupId)

		assert.EqualError(t, err, "failed to delete group: group id is empty")
	})
}
