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

var requestError = errors.New("request error")

func TestClientGroupCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Create Group Ok", func(t *testing.T) {
		expected := &model.Group{
			ID:   "test-id",
			Name: "test",
		}

		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		group, err := c.CreateGroup(context.Background(), "test")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, group)
	})
}

func TestClientGroupCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Group Error", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

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
		jsonResponse := `{
		  "data": {
		    "group": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		group, err := c.CreateGroup(context.Background(), "")

		assert.EqualError(t, err, "failed to create group: name is empty")
		assert.Nil(t, group)
	})
}

func TestClientGroupUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Ok", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		_, err := c.UpdateGroup(context.Background(), "groupId", "groupName")

		assert.NoError(t, err)
	})
}

func TestClientGroupUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Error", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		const groupId = "g1"
		_, err := c.UpdateGroup(context.Background(), groupId, "test")

		assert.EqualError(t, err, fmt.Sprintf("failed to update group with id %s: error_1", groupId))
	})
}

func TestClientGroupUpdateEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "ok": true,
		      "entity": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const groupId = "g1"
		_, err := c.UpdateGroup(context.Background(), groupId, "test")

		assert.EqualError(t, err, fmt.Sprintf("failed to update group with id %s: query result is empty", groupId))
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
		expected := &model.Group{
			ID:       "id",
			Name:     "name",
			Type:     "MANUAL",
			IsActive: true,
			UserIDs:  []string{},
		}

		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		group, err := c.ReadGroup(context.Background(), "id")

		assert.NoError(t, err)
		assert.Equal(t, expected, group)
	})
}

func TestClientGroupReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "group": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

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
		jsonResponse := `{
		  "data": {
		    "group": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		group, err := c.ReadGroup(context.Background(), "")

		assert.EqualError(t, err, "failed to read group: id is empty")
		assert.Nil(t, group)
	})
}

func TestClientGroupReadRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group - Request Error On Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "group": {
		      "id": "id",
		      "name": "name",
		      "type": "MANUAL",
		      "isActive": true,
		      "users": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": true
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "id-1",
		              "email": "user@1",
		              "role": "ADMIN"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewErrorResponder(requestError),
			),
		)

		const groupId = "g1"
		group, err := c.ReadGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id %s: Post "%s": %v`, groupId, c.GraphqlServerURL, requestError))
	})
}

func TestClientDeleteGroupOk(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Ok", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		err := c.DeleteGroup(context.Background(), "g1")

		assert.NoError(t, err)
	})
}

func TestClientDeleteEmptyGroupError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Empty Group Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "group": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		err := c.DeleteGroup(context.Background(), "")

		assert.EqualError(t, err, "failed to delete group: id is empty")
	})
}

func TestClientDeleteGroupError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Error", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

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
		expected := []*model.Group{
			{
				ID:      "id1",
				Name:    "group1",
				UserIDs: []string{},
			},
			{
				ID:      "id2",
				Name:    "group2",
				UserIDs: []string{},
			},
			{
				ID:      "id3",
				Name:    "group3",
				UserIDs: []string{},
			},
		}

		jsonResponse := `{
		  "data": {
		    "groups": {
		      "edges": [
		        {
		          "node": {
		            "id": "id1",
		            "name": "group1"
		          }
		        },
		        {
		          "node": {
		            "id": "id2",
		            "name": "group2"
		          }
		        },
		        {
		          "node": {
		            "id": "id3",
		            "name": "group3"
		          }
		        }
		      ]
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		groups, err := c.ReadGroups(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expected, groups)
	})
}

func TestClientGroupsReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups Error", func(t *testing.T) {
		emptyResponse := `{
		  "data": {
		    "groups": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, emptyResponse))

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

func TestClientGroupsReadRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups - Request Error on Fetching", func(t *testing.T) {
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
		            "id": "id1",
		            "name": "group1"
		          }
		        }
		      ]
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewErrorResponder(errors.New("error_1")),
			),
		)

		groups, err := c.ReadGroups(context.Background())

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: Post "%s": error_1`, c.GraphqlServerURL))
	})
}

func TestClientGroupsReadEmptyResultOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups - Empty Result on Fetching", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "groups": {
			"pageInfo": {
		        "endCursor": "cursor-001",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id1",
		            "name": "group1"
		          }
		        }
		      ]
		    }
		  }
		}`

		response2 := `{}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, response1),
				httpmock.NewStringResponder(200, response2),
			),
		)

		groups, err := c.ReadGroups(context.Background())

		assert.Nil(t, groups)
		assert.EqualError(t, err, `failed to read group with id All: query result is empty`)
	})
}

func TestClientGroupsReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups All - Ok", func(t *testing.T) {
		expected := []*model.Group{
			{ID: "id-1", Name: "group-1", UserIDs: []string{}},
			{ID: "id-2", Name: "group-2", UserIDs: []string{}},
			{ID: "id-3", Name: "group-3", UserIDs: []string{}},
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

		groups, err := c.ReadGroups(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expected, groups)
	})
}

func TestClientGroupsReadByNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Ok", func(t *testing.T) {
		expected := []*model.Group{
			{ID: "id-1", Name: "group-1", UserIDs: []string{}},
			{ID: "id-2", Name: "group-2", UserIDs: []string{}},
			{ID: "id-3", Name: "group-3", UserIDs: []string{}},
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

		assert.NoError(t, err)
		assert.Equal(t, expected, groups)
	})
}

func TestClientGroupsReadByNameRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Request Error on Fetching", func(t *testing.T) {
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
		        }
		      ]
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewErrorResponder(errors.New("error_1")),
			),
		)

		groups, err := c.ReadGroupsByName(context.Background(), "group-1-2-3")

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: Post "%s": error_1`, c.GraphqlServerURL))
	})
}

func TestClientGroupsReadByNameEmptyResultOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Empty Result on Fetching", func(t *testing.T) {
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
		        }
		      ]
		    }
		  }
		}`

		emptyResponse := `{}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewStringResponder(200, emptyResponse),
			),
		)

		groups, err := c.ReadGroupsByName(context.Background(), "group-1-2-3")

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: query result is empty`))
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

		groups, err := c.ReadGroupsByName(context.Background(), "")

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
				filter:           &client.GroupsFilter{Name: optionalString("Group")},
				expectedGroupIds: []string{"g1", "g2"},
			},
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

func TestClientFilterGroupsEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Filter Groups - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groups": {
		      "edges": []
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		groups, err := c.FilterGroups(context.Background(), nil)

		assert.NoError(t, err)
		assert.Nil(t, groups)
	})
}

func TestClientAssignGroupUsersOk(t *testing.T) {
	t.Run("Test Twingate Resource : Assign Group Users - Ok", func(t *testing.T) {
		expected := &model.Group{
			ID:       "id-1",
			Name:     "group-1",
			Type:     "MANUAL",
			IsActive: true,
			UserIDs:  []string{"id-1", "id-2"},
		}

		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": {
		        "id": "id-1",
		        "name": "group-1",
		        "isActive": true,
		        "type": "MANUAL",
		        "users": {
		          "pageInfo": {
		            "endCursor": "cursor-001",
		            "hasNextPage": true
		          },
		          "edges": [
		            {
		              "node": {
		                "id": "id-1",
		                "email": "user@1",
		                "role": "ADMIN"
		              }
		            }
		          ]
		        }
		      },
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		nextPage := `{
		  "data": {
		    "group": {
		      "id": "id-1",
		      "name": "group-1",
		      "isActive": true,
		      "type": "MANUAL",
		      "users": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "id-2",
		              "email": "user@2",
		              "role": "DEVOPS"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewStringResponder(http.StatusOK, nextPage),
			),
		)

		group, err := c.AssignGroupUsers(context.Background(), "id-1", []string{"id-1", "id-2"})

		assert.NoError(t, err)
		assert.Equal(t, expected, group)
	})
}

func TestClientAssignGroupUsersEmptyGroupID(t *testing.T) {
	t.Run("Test Twingate Resource : Assign Group Users - Empty Group ID", func(t *testing.T) {
		c := newHTTPMockClient()
		group, err := c.AssignGroupUsers(context.Background(), "", nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, "failed to update group: id is empty")
	})
}

func TestClientAssignGroupUsersWithNilUsersIDs(t *testing.T) {
	t.Run("Test Twingate Resource : Assign Group Users - With Nil User IDs", func(t *testing.T) {
		expected := &model.Group{
			ID:       "id-1",
			Name:     "group-1",
			Type:     "MANUAL",
			IsActive: true,
			UserIDs:  []string{},
		}

		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": {
		        "id": "id-1",
		        "name": "group-1",
		        "isActive": true,
		        "type": "MANUAL",
		        "users": {
		          "edges": []
		        }
		      },
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		group, err := c.AssignGroupUsers(context.Background(), "id-1", nil)

		assert.NoError(t, err)
		assert.Equal(t, expected, group)
	})
}

func TestClientAssignGroupUsersRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Assign Group Users - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(requestError))

		group, err := c.AssignGroupUsers(context.Background(), "id-1", nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to update group with id id-1: Post "%s": %v`, c.GraphqlServerURL, requestError))
	})
}

func TestClientAssignGroupUsersResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Assign Group Users - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": null,
		      "ok": false,
		      "error": "response error"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		group, err := c.AssignGroupUsers(context.Background(), "id-1", nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, `failed to update group with id id-1: response error`)
	})
}

func TestClientAssignGroupUsersEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Assign Group Users - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": null,
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		group, err := c.AssignGroupUsers(context.Background(), "id-1", nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, `failed to update group with id id-1: query result is empty`)
	})
}

func TestClientAssignGroupUsersRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Assign Group Users - Request Error On Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": {
		        "id": "id-1",
		        "name": "group-1",
		        "isActive": true,
		        "type": "MANUAL",
		        "users": {
		          "pageInfo": {
		            "endCursor": "cursor-001",
		            "hasNextPage": true
		          },
		          "edges": [
		            {
		              "node": {
		                "id": "id-1",
		                "email": "user@1",
		                "role": "ADMIN"
		              }
		            }
		          ]
		        }
		      },
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewErrorResponder(requestError),
			),
		)

		group, err := c.AssignGroupUsers(context.Background(), "id-1", []string{"id-1", "id-2"})

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id id-1: Post "%s": %v`, c.GraphqlServerURL, requestError))
	})
}

func TestClientAssignGroupUsersEmptyResponseOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Assign Group Users - Empty Response On Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": {
		        "id": "id-1",
		        "name": "group-1",
		        "isActive": true,
		        "type": "MANUAL",
		        "users": {
		          "pageInfo": {
		            "endCursor": "cursor-001",
		            "hasNextPage": true
		          },
		          "edges": [
		            {
		              "node": {
		                "id": "id-1",
		                "email": "user@1",
		                "role": "ADMIN"
		              }
		            }
		          ]
		        }
		      },
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		nextPage := `{
		  "data": {
		    "group": {
		      "id": "id-1",
		      "name": "group-1",
		      "isActive": true,
		      "type": "MANUAL",
		      "users": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": false
		        },
		        "edges": []
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewStringResponder(http.StatusOK, nextPage),
			),
		)

		group, err := c.AssignGroupUsers(context.Background(), "id-1", []string{"id-1", "id-2"})

		assert.Nil(t, group)
		assert.EqualError(t, err, `failed to read group with id id-1: query result is empty`)
	})
}

func TestClientRemoveGroupUsersOk(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Group Users - Ok", func(t *testing.T) {
		expected := &model.Group{
			ID:       "id-1",
			Name:     "group-1",
			Type:     "MANUAL",
			IsActive: true,
			UserIDs:  []string{"id-3", "id-4"},
		}

		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": {
		        "id": "id-1",
		        "name": "group-1",
		        "isActive": true,
		        "type": "MANUAL",
		        "users": {
		          "pageInfo": {
		            "endCursor": "cursor-001",
		            "hasNextPage": true
		          },
		          "edges": [
		            {
		              "node": {
		                "id": "id-3",
		                "email": "user@3",
		                "role": "ADMIN"
		              }
		            }
		          ]
		        }
		      },
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		nextPage := `{
		  "data": {
		    "group": {
		      "id": "id-1",
		      "name": "group-1",
		      "isActive": true,
		      "type": "MANUAL",
		      "users": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "id-4",
		              "email": "user@4",
		              "role": "DEVOPS"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewStringResponder(http.StatusOK, nextPage),
			),
		)

		group, err := c.RemoveGroupUsers(context.Background(), "id-1", []string{"id-1", "id-2"})

		assert.NoError(t, err)
		assert.Equal(t, expected, group)
	})
}

func TestClientRemoveGroupUsersEmptyGroupID(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Group Users - Empty Group ID", func(t *testing.T) {
		c := newHTTPMockClient()
		group, err := c.RemoveGroupUsers(context.Background(), "", nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, "failed to update group: id is empty")
	})
}

func TestClientRemoveGroupUsersWithNilUsersIDs(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Group Users - With Nil User IDs", func(t *testing.T) {
		expected := &model.Group{
			ID:       "id-1",
			Name:     "group-1",
			Type:     "MANUAL",
			IsActive: true,
			UserIDs:  []string{},
		}

		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": {
		        "id": "id-1",
		        "name": "group-1",
		        "isActive": true,
		        "type": "MANUAL",
		        "users": {
		          "edges": []
		        }
		      },
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		group, err := c.RemoveGroupUsers(context.Background(), "id-1", nil)

		assert.NoError(t, err)
		assert.Equal(t, expected, group)
	})
}

func TestClientRemoveGroupUsersRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Group Users - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(requestError))

		group, err := c.RemoveGroupUsers(context.Background(), "id-1", nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to update group with id id-1: Post "%s": %v`, c.GraphqlServerURL, requestError))
	})
}

func TestClientRemoveGroupUsersResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Group Users - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": null,
		      "ok": false,
		      "error": "response error"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		group, err := c.RemoveGroupUsers(context.Background(), "id-1", nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, `failed to update group with id id-1: response error`)
	})
}

func TestClientRemoveGroupUsersEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Group Users - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": null,
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		group, err := c.RemoveGroupUsers(context.Background(), "id-1", nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, `failed to update group with id id-1: query result is empty`)
	})
}

func TestClientRemoveGroupUsersRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Group Users - Request Error On Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": {
		        "id": "id-1",
		        "name": "group-1",
		        "isActive": true,
		        "type": "MANUAL",
		        "users": {
		          "pageInfo": {
		            "endCursor": "cursor-001",
		            "hasNextPage": true
		          },
		          "edges": [
		            {
		              "node": {
		                "id": "id-1",
		                "email": "user@1",
		                "role": "ADMIN"
		              }
		            }
		          ]
		        }
		      },
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewErrorResponder(requestError),
			),
		)

		group, err := c.RemoveGroupUsers(context.Background(), "id-1", []string{"id-1", "id-2"})

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id id-1: Post "%s": %v`, c.GraphqlServerURL, requestError))
	})
}

func TestClientRemoveGroupUsersEmptyResponseOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Group Users - Empty Response On Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "groupUpdate": {
		      "entity": {
		        "id": "id-1",
		        "name": "group-1",
		        "isActive": true,
		        "type": "MANUAL",
		        "users": {
		          "pageInfo": {
		            "endCursor": "cursor-001",
		            "hasNextPage": true
		          },
		          "edges": [
		            {
		              "node": {
		                "id": "id-1",
		                "email": "user@1",
		                "role": "ADMIN"
		              }
		            }
		          ]
		        }
		      },
		      "ok": true,
		      "error": ""
		    }
		  }
		}`

		nextPage := `{
		  "data": {
		    "group": {
		      "id": "id-1",
		      "name": "group-1",
		      "isActive": true,
		      "type": "MANUAL",
		      "users": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": false
		        },
		        "edges": []
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewStringResponder(http.StatusOK, nextPage),
			),
		)

		group, err := c.RemoveGroupUsers(context.Background(), "id-1", []string{"id-1", "id-2"})

		assert.Nil(t, group)
		assert.EqualError(t, err, `failed to read group with id id-1: query result is empty`)
	})
}
