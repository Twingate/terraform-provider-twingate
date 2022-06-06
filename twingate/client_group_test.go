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

func TestClientGroupCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Create Group Ok", func(t *testing.T) {
		const (
			groupName = "test"
			groupID   = "test-id"
		)

		// response JSON
		createGroupOkJson := fmt.Sprintf(`{
		  "data": {
			"groupCreate": {
			  "entity": {
				"id": "%s",
				"name": "%s"
			  },
			  "ok": true,
			  "error": null
			}
		  }
		}`, groupID, groupName)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createGroupOkJson))

		group, err := client.createGroup(context.Background(), groupName)

		assert.Nil(t, err)
		assert.EqualValues(t, groupID, group.ID)
		assert.EqualValues(t, groupName, group.Name)
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

		group, err := client.createGroup(context.Background(), "test")

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

		group, err := client.createGroup(context.Background(), "test")

		assert.EqualError(t, err, fmt.Sprintf(`failed to create group: Message: Post "%s": error_1, Locations: []`, client.GraphqlServerURL))
		assert.Nil(t, group)
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

		group, err := client.createGroup(context.Background(), "")

		assert.EqualError(t, err, "failed to create group: name is empty")
		assert.Nil(t, group)
	})
}

func TestClientGroupUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Ok", func(t *testing.T) {
		const (
			groupID   = "id"
			groupName = "test"
		)

		// response JSON
		updateGroupOkJson := fmt.Sprintf(`{
		  "data": {
			"groupUpdate": {
			  "entity": {
				"id": "%s",
				"name": "%s"
			  },
			  "ok": true,
			  "error": null
			}
		  }
		}`, groupID, groupName)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, updateGroupOkJson))

		err := client.updateGroup(context.Background(), groupID, groupName)

		assert.Nil(t, err)
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

		const groupId = "g1"
		err := client.updateGroup(context.Background(), groupId, "test")

		assert.EqualError(t, err, fmt.Sprintf("failed to update group with id %s: error_1", groupId))
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

		const groupId = "g1"
		err := client.updateGroup(context.Background(), groupId, "test")

		assert.EqualError(t, err, fmt.Sprintf(`failed to update group with id %s: Message: Post "%s": error_1, Locations: []`, groupId, client.GraphqlServerURL))
	})
}

func TestClientGroupUpdateWithEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group With Empty Name", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		err := client.updateGroup(context.Background(), "id", "")

		assert.EqualError(t, err, "failed to update group: name is empty")
	})
}

func TestClientGroupUpdateWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group With Empty ID", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		err := client.updateGroup(context.Background(), "", "name")

		assert.EqualError(t, err, "failed to update group: id is empty")
	})
}

func TestClientGroupReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Ok", func(t *testing.T) {
		const (
			groupID   = "id"
			groupName = "name"
		)

		// response JSON
		readGroupOkJson := fmt.Sprintf(`{
		  "data": {
			"group": {
			  "id": "%s",
			  "name": "%s"
			}
		  }
		}`, groupID, groupName)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		group, err := client.readGroup(context.Background(), groupID)

		assert.Nil(t, err)
		assert.NotNil(t, group)
		assert.EqualValues(t, groupID, group.ID)
		assert.EqualValues(t, groupName, group.Name)
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
		const groupId = "g1"

		group, err := client.readGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf("failed to read group with id %s", groupId))
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
		const groupId = "g1"

		group, err := client.readGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id %s: Message: Post "%s": error_1, Locations: []`, groupId, client.GraphqlServerURL))
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

		group, err := client.readGroup(context.Background(), "")

		assert.EqualError(t, err, "failed to read group: id is empty")
		assert.Nil(t, group)
	})
}

func TestClientDeleteGroupOk(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Ok", func(t *testing.T) {
		// response JSON
		deleteGroupOkJson := `{
		  "data": {
			"groupDelete": {
			  "ok": true,
			  "error": null
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteGroupOkJson))

		err := client.deleteGroup(context.Background(), "g1")

		assert.Nil(t, err)
	})
}

func TestClientDeleteEmptyGroupError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Empty Group Error", func(t *testing.T) {
		// response JSON
		deleteGroupOkJson := `{
		  "data": {
			"group": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteGroupOkJson))

		err := client.deleteGroup(context.Background(), "")

		assert.EqualError(t, err, "failed to delete group: id is empty")
	})
}

func TestClientDeleteGroupError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Error", func(t *testing.T) {
		// response JSON
		deleteGroupOkJson := `{
		  "data": {
			"groupDelete": {
			  "ok": false,
			  "error": "error_1"
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteGroupOkJson))
		const groupID = "g1"

		err := client.deleteGroup(context.Background(), groupID)

		assert.EqualError(t, err, fmt.Sprintf("failed to delete group with id %s: error_1", groupID))
	})
}

func TestClientDeleteGroupRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Request Error", func(t *testing.T) {
		// response JSON
		deleteGroupOkJson := `{
		  "data": {
			"groupDelete": {
			  "ok": false,
			  "error": "error_1"
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, deleteGroupOkJson)
				return resp, errors.New("error_2")
			})
		const groupID = "g1"

		err := client.deleteGroup(context.Background(), groupID)

		assert.EqualError(t, err, fmt.Sprintf(`failed to delete group with id %s: Message: Post "%s": error_2, Locations: []`, groupID, client.GraphqlServerURL))
	})
}

func TestClientGroupsReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups Ok", func(t *testing.T) {
		ids := []string{"id1", "id2", "id3"}
		names := []string{"group1", "group2", "group3"}
		// response JSON
		readGroupOkJson := fmt.Sprintf(`{
		  "data": {
			"groups": {
			  "edges": [
				{
				  "node": {
					"id": "%s",
					"name": "%s"
				  }
				},
				{
				  "node": {
					"id": "%s",
					"name": "%s"
				  }
				},
				{
				  "node": {
					"id": "%s",
					"name": "%s"
				  }
				}
			  ]
			}
		  }
		}`, ids[0], names[0], ids[1], names[1], ids[2], names[2])

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		groups, err := client.readGroups(context.Background())

		assert.Nil(t, err)
		assert.NotNil(t, groups)
		assert.EqualValues(t, len(ids), len(groups))
		for i, id := range ids {
			assert.EqualValues(t, id, groups[i].ID)
			assert.EqualValues(t, names[i], groups[i].Name)
		}
	})
}

func TestClientGroupsReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups Error", func(t *testing.T) {
		// response JSON
		readGroupOkJson := `{
		  "data": {
			"groups": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		groups, err := client.readGroups(context.Background())

		assert.Nil(t, groups)
		assert.EqualError(t, err, "failed to read group with id All")
	})
}

func TestClientGroupsReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups Request Error", func(t *testing.T) {
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

		group, err := client.readGroups(context.Background())

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: Message: Post "%s": error_1, Locations: []`, client.GraphqlServerURL))
	})
}
