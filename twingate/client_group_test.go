package twingate

import (
	"context"
	"errors"
	"fmt"
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
				"id": "test-id",
				"name": "test"
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
		assert.EqualValues(t, "test", group.Name)
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

		assert.EqualError(t, err, fmt.Sprintf(`failed to create group: Post "%s": error_1`, client.GraphqlServerURL))
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
		groupName := graphql.String("")

		group, err := client.createGroup(context.Background(), groupName)

		assert.EqualError(t, err, "failed to create group: name is empty")
		assert.Nil(t, group)
	})
}

func TestClientGroupUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Ok", func(t *testing.T) {
		// response JSON
		updateGroupOkJson := `{
		  "data": {
			"groupUpdate": {
			  "entity": {
				"id": "id",
				"name": "test"
			  },
			  "ok": true,
			  "error": null
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
		groupName := graphql.String("test")
		groupId := graphql.ID("g1")
		err := client.updateGroup(context.Background(), groupId, groupName)

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

		groupName := graphql.String("test")
		groupId := graphql.ID("g1")
		err := client.updateGroup(context.Background(), groupId, groupName)

		assert.EqualError(t, err, fmt.Sprintf(`failed to update group with id %s: Post "%s": error_1`, groupId, client.GraphqlServerURL))
	})
}

func TestClientGroupUpdateWithEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group With Empty Name", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		groupName := graphql.String("")
		groupId := graphql.ID("id")
		err := client.updateGroup(context.Background(), groupId, groupName)

		assert.EqualError(t, err, "failed to update group: name is empty")
	})
}

func TestClientGroupUpdateWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group With Empty ID", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		groupName := graphql.String("name")
		groupId := graphql.ID("")
		err := client.updateGroup(context.Background(), groupId, groupName)

		assert.EqualError(t, err, "failed to update group: id is empty")
	})
}

func TestClientGroupReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Ok", func(t *testing.T) {
		// response JSON
		readGroupOkJson := `{
		  "data": {
			"group": {
			  "id": "id",
			  "name": "name"
			}
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))
		groupId := graphql.ID("id")

		group, err := client.readGroup(context.Background(), groupId)

		assert.Nil(t, err)
		assert.NotNil(t, group)
		assert.EqualValues(t, groupId, group.ID)
		assert.EqualValues(t, "name", group.Name)
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
		groupId := graphql.ID("g1")

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
		groupId := graphql.ID("g1")

		group, err := client.readGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id %s: Post "%s": error_1`, groupId, client.GraphqlServerURL))
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
		groupId := graphql.ID("g1")

		err := client.deleteGroup(context.Background(), groupId)

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
		groupId := graphql.ID("")

		err := client.deleteGroup(context.Background(), groupId)

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
		groupId := graphql.ID("g1")

		err := client.deleteGroup(context.Background(), groupId)

		assert.EqualError(t, err, "failed to delete group with id g1: error_1")
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
		groupId := graphql.ID("g1")

		err := client.deleteGroup(context.Background(), groupId)

		assert.EqualError(t, err, fmt.Sprintf(`failed to delete group with id g1: Post "%s": error_2`, client.GraphqlServerURL))
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
		assert.EqualError(t, err, "failed to read group with id All: query result is empty")
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
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: Post "%s": error_1`, client.GraphqlServerURL))
	})
}

func TestClientGroupsReadByNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Ok", func(t *testing.T) {
		const groupName = "group-1"
		ids := []string{"id-1", "id-2"}
		jsonResponse := fmt.Sprintf(`{
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
				}
			  ]
			}
		  }
		}`, ids[0], groupName, ids[1], groupName)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		groups, err := client.readGroupsByName(context.Background(), groupName)

		assert.Nil(t, err)
		assert.NotNil(t, groups)
		assert.Len(t, groups, len(ids))
		for i, id := range ids {
			assert.EqualValues(t, id, groups[i].ID)
			assert.EqualValues(t, groupName, groups[i].Name)
		}
	})
}

func TestClientGroupsReadByNameEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Empty Result", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
			"groups": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const groupName = "group-name"
		groups, err := client.readGroupsByName(context.Background(), groupName)

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf("failed to read group with name %s: query result is empty", groupName))
	})
}

func TestClientGroupsReadByNameRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
			"groups": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, jsonResponse)
				return resp, errors.New("error_1")
			})

		const groupName = "group-name"
		groups, err := client.readGroupsByName(context.Background(), groupName)

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with name %s: Post "%s": error_1`, groupName, client.GraphqlServerURL))
	})
}

func TestClientGroupsReadByNameErrorEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Error Empty Name", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
			"groups": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const groupName = ""
		groups, err := client.readGroupsByName(context.Background(), groupName)

		assert.Nil(t, groups)
		assert.EqualError(t, err, "failed to read group: group name is empty")
	})
}

func TestClientFilterGroupsOk(t *testing.T) {
	t.Run("Test Twingate Resource : Filter Groups - Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
			"groups": {
			  "edges": [
				{
				  "node": {
					"id": "g1",
					"name": "Group 1",
					"type": "MANUAL",
					"isActive": true
				  }
				},
				{
				  "node": {
					"id": "g2",
					"name": "Group 2",
					"type": "SYSTEM",
					"isActive": false
				  }
				}
			  ]
			}
		  }
		}`

		testData := []struct {
			filter           *GroupsFilter
			expectedGroupIds []string
		}{
			{
				filter:           &GroupsFilter{Type: optionalString("MANUAL")},
				expectedGroupIds: []string{"g1"},
			},
			{
				filter:           &GroupsFilter{Type: optionalString("SYSTEM")},
				expectedGroupIds: []string{"g2"},
			},
			{
				filter: &GroupsFilter{Type: optionalString("SYNCED")},
			},
			{
				filter:           &GroupsFilter{IsActive: optionalBool(true)},
				expectedGroupIds: []string{"g1"},
			},
			{
				filter:           &GroupsFilter{IsActive: optionalBool(false)},
				expectedGroupIds: []string{"g2"},
			},
			{
				filter:           &GroupsFilter{Type: optionalString("SYSTEM"), IsActive: optionalBool(false)},
				expectedGroupIds: []string{"g2"},
			},
			{
				filter:           &GroupsFilter{Type: optionalString("MANUAL"), IsActive: optionalBool(true)},
				expectedGroupIds: []string{"g1"},
			},
			{
				filter: &GroupsFilter{Type: optionalString("MANUAL"), IsActive: optionalBool(false)},
			},
			{
				expectedGroupIds: []string{"g1", "g2"},
			},
		}

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		for _, td := range testData {
			groups, err := client.filterGroups(context.Background(), td.filter)

			assert.Nil(t, err)
			assert.Len(t, groups, len(td.expectedGroupIds))
			if td.expectedGroupIds == nil {
				assert.Nil(t, groups)
			} else {
				assert.NotNil(t, groups)
				for i, id := range td.expectedGroupIds {
					assert.EqualValues(t, id, groups[i].ID)
				}
			}
		}
	})
}

func optionalString(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

func optionalBool(val bool) *bool {
	return &val
}

func TestClientFilterGroupsRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Filter Groups - Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
			"groups": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, jsonResponse)
				return resp, errors.New("error_1")
			})

		groups, err := client.filterGroups(context.Background(), nil)

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: Post "%s": error_1`, client.GraphqlServerURL))
	})
}
