package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, createGroupOkJson))

		group, err := c.CreateGroup(context.Background(), "test")

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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, createGroupOkJson))

		group, err := c.CreateGroup(context.Background(), "test")

		assert.EqualError(t, err, "failed to create group: error_1")
		assert.Nil(t, group)
	})
}

func TestClientGroupCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Group Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		group, err := c.CreateGroup(context.Background(), "test")

		assert.EqualError(t, err, fmt.Sprintf(`failed to create group: Post "%s": error_1`, c.GraphqlServerURL))
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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		group, err := c.CreateGroup(context.Background(), "")

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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, updateGroupOkJson))

		_, err := c.UpdateGroup(context.Background(), "groupId", "groupName")

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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, updateGroupOkJson))

		const groupId = "g1"
		_, err := c.UpdateGroup(context.Background(), groupId, "test")

		assert.EqualError(t, err, fmt.Sprintf("failed to update group with id %s: error_1", groupId))
	})
}

func TestClientGroupUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		const groupId = "g1"
		_, err := c.UpdateGroup(context.Background(), groupId, "test")

		assert.EqualError(t, err, fmt.Sprintf(`failed to update group with id %s: Post "%s": error_1`, groupId, c.GraphqlServerURL))
	})
}

func TestClientGroupUpdateWithEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group With Empty Name", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		_, err := c.UpdateGroup(context.Background(), "id", "")

		assert.EqualError(t, err, "failed to update group: name is empty")
	})
}

func TestClientGroupUpdateWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		_, err := c.UpdateGroup(context.Background(), "", "groupName")

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
		      "name": "name",
		      "type": "MANUAL",
		      "isActive": true
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		const groupId = "id"
		group, err := c.ReadGroup(context.Background(), groupId)

		assert.Nil(t, err)
		assert.NotNil(t, group)
		assert.EqualValues(t, groupId, group.ID)
		assert.EqualValues(t, "name", group.Name)
		assert.EqualValues(t, "MANUAL", group.Type)
		assert.EqualValues(t, true, group.IsActive)
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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		const groupId = "g1"
		group, err := c.ReadGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf("failed to read group with id %s: query result is empty", groupId))
	})
}

func TestClientGroupReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		const groupId = "g1"
		group, err := c.ReadGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id %s: Post "%s": error_1`, groupId, c.GraphqlServerURL))
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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		group, err := c.ReadGroup(context.Background(), "")

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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteGroupOkJson))

		err := c.DeleteGroup(context.Background(), "g1")

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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteGroupOkJson))

		err := c.DeleteGroup(context.Background(), "")

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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteGroupOkJson))

		err := c.DeleteGroup(context.Background(), "g1")

		assert.EqualError(t, err, "failed to delete group with id g1: error_1")
	})
}

func TestClientDeleteGroupRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_2")))

		err := c.DeleteGroup(context.Background(), "g1")

		assert.EqualError(t, err, fmt.Sprintf(`failed to delete group with id g1: Post "%s": error_2`, c.GraphqlServerURL))
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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		groups, err := c.ReadGroups(context.Background())

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

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, readGroupOkJson))

		groups, err := c.ReadGroups(context.Background())

		assert.Nil(t, groups)
		assert.EqualError(t, err, "failed to read group with id All: query result is empty")
	})
}

func TestClientGroupsReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		group, err := c.ReadGroups(context.Background())

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: Post "%s": error_1`, c.GraphqlServerURL))
	})
}

func TestClientGroupsReadByNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Ok", func(t *testing.T) {
		expected := []*model.Group{
			{ID: "id-1", Name: "group-1"},
			{ID: "id-2", Name: "group-2"},
			{ID: "id-3", Name: "group-3"},
		}

		jsonResponse := `{
		  "data": {
		    "groups": {
		      "pageInfo": {
		        "endCursor": "cursor-001",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "group-1"
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "group-2"
		          }
		        }
		      ]
		    }
		  }
		}`

		nextPage := `{
		  "data": {
		    "groups": {
		      "pageInfo": {
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-3",
		            "name": "group-3"
		          }
		        }
		      ]
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.ResponderFromMultipleResponses(
				[]*http.Response{
					httpmock.NewStringResponse(200, jsonResponse),
					httpmock.NewStringResponse(200, nextPage),
				}),
		)

		groups, err := c.ReadGroupsByName(context.Background(), "group-1-2-3")

		assert.Nil(t, err)
		assert.Equal(t, expected, groups)
	})
}

func TestClientGroupsReadByNameEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Empty Result", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groups": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const groupName = "group-name"
		groups, err := c.ReadGroupsByName(context.Background(), groupName)

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf("failed to read group with name %s: query result is empty", groupName))
	})
}

func TestClientGroupsReadByNameRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		const groupName = "group-name"
		groups, err := c.ReadGroupsByName(context.Background(), groupName)

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with name %s: Post "%s": error_1`, groupName, c.GraphqlServerURL))
	})
}

func TestClientGroupsReadByNameErrorEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Error Empty Name", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groups": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const groupName = ""
		groups, err := c.ReadGroupsByName(context.Background(), groupName)

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
			filter           *client.GroupsFilter
			expectedGroupIds []string
		}{
			{
				filter:           &client.GroupsFilter{Type: optionalString("MANUAL")},
				expectedGroupIds: []string{"g1"},
			},
			{
				filter:           &client.GroupsFilter{Type: optionalString("SYSTEM")},
				expectedGroupIds: []string{"g2"},
			},
			{
				filter: &client.GroupsFilter{Type: optionalString("SYNCED")},
			},
			{
				filter:           &client.GroupsFilter{IsActive: optionalBool(true)},
				expectedGroupIds: []string{"g1"},
			},
			{
				filter:           &client.GroupsFilter{IsActive: optionalBool(false)},
				expectedGroupIds: []string{"g2"},
			},
			{
				filter:           &client.GroupsFilter{Type: optionalString("SYSTEM"), IsActive: optionalBool(false)},
				expectedGroupIds: []string{"g2"},
			},
			{
				filter:           &client.GroupsFilter{Type: optionalString("MANUAL"), IsActive: optionalBool(true)},
				expectedGroupIds: []string{"g1"},
			},
			{
				filter: &client.GroupsFilter{Type: optionalString("MANUAL"), IsActive: optionalBool(false)},
			},
			{
				expectedGroupIds: []string{"g1", "g2"},
			},
		}

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		for _, td := range testData {
			groups, err := c.FilterGroups(context.Background(), td.filter)

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
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		groups, err := c.FilterGroups(context.Background(), nil)

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: Post "%s": error_1`, c.GraphqlServerURL))
	})
}
