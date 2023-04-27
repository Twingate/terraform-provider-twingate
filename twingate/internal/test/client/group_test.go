package client

import (
	"context"
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

		group, err := c.CreateGroup(context.Background(), &model.Group{Name: "test"})

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

		group, err := c.CreateGroup(context.Background(), &model.Group{Name: "test"})

		assert.EqualError(t, err, "failed to create group with name test: error_1")
		assert.Nil(t, group)
	})
}

func TestClientGroupCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Group Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		group, err := c.CreateGroup(context.Background(), &model.Group{Name: "test"})

		assert.EqualError(t, err, graphqlErr(c, "failed to create group with name test", errBadRequest))
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

		group, err := c.CreateGroup(context.Background(), nil)

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

		_, err := c.UpdateGroup(context.Background(), &model.Group{ID: "groupId", Name: "groupName"})

		assert.NoError(t, err)
	})
}

func TestClientGroupUpdateErrorOnFetchPages(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Error On Fetch Pages", func(t *testing.T) {
		jsonResponse := `{
          "data": {
            "groupUpdate": {
              "entity": {
                "id": "group-id",
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
                        "id": "user-id"
                      }
                    }
                  ]
                }
              },
              "ok": true,
              "error": null
            }
          }
        }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewErrorResponder(errBadRequest),
			))

		_, err := c.UpdateGroup(context.Background(), &model.Group{ID: "group-id", Name: "groupName"})

		assert.EqualError(t, err, graphqlErr(c, "failed to read group with id group-id", errBadRequest))
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
		_, err := c.UpdateGroup(context.Background(), &model.Group{ID: groupId, Name: "test"})

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
		_, err := c.UpdateGroup(context.Background(), &model.Group{ID: groupId, Name: "test"})

		assert.EqualError(t, err, fmt.Sprintf("failed to update group with id %s: query result is empty", groupId))
	})
}

func TestClientGroupUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		const groupId = "g1"
		_, err := c.UpdateGroup(context.Background(), &model.Group{ID: groupId, Name: "test"})

		assert.EqualError(t, err, graphqlErr(c, "failed to update group with id "+groupId, errBadRequest))
	})
}

func TestClientGroupUpdateWithEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group With Empty Name", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		_, err := c.UpdateGroup(context.Background(), &model.Group{ID: "id"})

		assert.EqualError(t, err, "failed to update group: name is empty")
	})
}

func TestClientGroupUpdateWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update Group With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		_, err := c.UpdateGroup(context.Background(), &model.Group{Name: "test"})

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
			Users:    []string{},
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
			httpmock.NewErrorResponder(errBadRequest))

		const groupId = "g1"
		group, err := c.ReadGroup(context.Background(), groupId)

		assert.Nil(t, group)
		assert.EqualError(t, err, graphqlErr(c, "failed to read group with id "+groupId, errBadRequest))
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

func TestClientGroupReadErrorOnFetchPages(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Error On Fetch Pages", func(t *testing.T) {
		jsonResponse := `{
          "data": {
            "group": {
              "id": "group-id",
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
                      "id": "user-id"
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
				httpmock.NewErrorResponder(errBadRequest),
			))

		group, err := c.ReadGroup(context.Background(), "group-id")

		assert.Nil(t, group)
		assert.EqualError(t, err, graphqlErr(c, "failed to read group with id group-id", errBadRequest))
	})
}

func TestClientGroupReadEmptyOnFetchPages(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Error On Fetch Pages", func(t *testing.T) {
		response1 := `{
          "data": {
            "group": {
              "id": "group-id",
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
                      "id": "user-id"
                    }
                  }
                ]
              }
            }
          }
        }`

		response2 := `{
          "data": {
            "group": null
          }
        }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
			))

		group, err := c.ReadGroup(context.Background(), "group-id")

		assert.Nil(t, group)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id group-id: query result is empty`))
	})
}

func TestClientGroupReadOkOnFetchPages(t *testing.T) {
	t.Run("Test Twingate Resource : Read Group Error On Fetch Pages", func(t *testing.T) {
		expected := &model.Group{
			ID:       "group-id",
			Name:     "name",
			Type:     "MANUAL",
			IsActive: true,
			Users:    []string{"user-1", "user-2"},
		}

		response1 := `{
          "data": {
            "group": {
              "id": "group-id",
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
                      "id": "user-1"
                    }
                  }
                ]
              }
            }
          }
        }`

		response2 := `{
          "data": {
            "group": {
              "id": "group-id",
              "name": "name",
              "type": "MANUAL",
              "isActive": true,
              "users": {
                "pageInfo": {
                  "endCursor": "cursor-002",
                  "hasNextPage": false
                },
                "edges": [
                  {
                    "node": {
                      "id": "user-2"
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
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
			))

		group, err := c.ReadGroup(context.Background(), "group-id")

		assert.NoError(t, err)
		assert.Equal(t, expected, group)
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
			httpmock.NewErrorResponder(errBadRequest))

		err := c.DeleteGroup(context.Background(), "g1")

		assert.EqualError(t, err, graphqlErr(c, "failed to delete group with id g1", errBadRequest))
	})
}

func TestClientGroupsReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups Ok", func(t *testing.T) {
		expected := []*model.Group{
			{
				ID:    "id1",
				Name:  "group1",
				Users: []string{},
			},
			{
				ID:    "id2",
				Name:  "group2",
				Users: []string{},
			},
			{
				ID:    "id3",
				Name:  "group3",
				Users: []string{},
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

		groups, err := c.ReadGroups(context.Background(), nil)

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

		groups, err := c.ReadGroups(context.Background(), nil)

		assert.Nil(t, groups)
		assert.EqualError(t, err, "failed to read group with id All: query result is empty")
	})
}

func TestClientGroupsReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		group, err := c.ReadGroups(context.Background(), nil)

		assert.Nil(t, group)
		assert.EqualError(t, err, graphqlErr(c, "failed to read group with id All", errBadRequest))
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
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		groups, err := c.ReadGroups(context.Background(), nil)

		assert.Nil(t, groups)
		assert.EqualError(t, err, graphqlErr(c, "failed to read group with id All", errBadRequest))
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

		groups, err := c.ReadGroups(context.Background(), nil)

		assert.Nil(t, groups)
		assert.EqualError(t, err, `failed to read group with id All: query result is empty`)
	})
}

func TestClientGroupsReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups All - Ok", func(t *testing.T) {
		expected := []*model.Group{
			{ID: "id-1", Name: "group-1", Users: []string{}},
			{ID: "id-2", Name: "group-2", Users: []string{}},
			{ID: "id-3", Name: "group-3", Users: []string{}},
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

		groups, err := c.ReadGroups(context.Background(), nil)

		assert.NoError(t, err)
		assert.Equal(t, expected, groups)
	})
}

func TestClientGroupsReadByNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Ok", func(t *testing.T) {
		expected := []*model.Group{
			{ID: "id-1", Name: "group-1", Users: []string{}},
			{ID: "id-2", Name: "group-2", Users: []string{}},
			{ID: "id-3", Name: "group-3", Users: []string{}},
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

		groups, err := c.ReadGroups(context.Background(), &model.GroupsFilter{Name: optionalString("group-1-2-3")})

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
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		groups, err := c.ReadGroups(context.Background(), &model.GroupsFilter{Name: optionalString("group-1-2-3")})

		assert.Nil(t, groups)
		assert.EqualError(t, err, graphqlErr(c, "failed to read group with id All", errBadRequest))
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

		groups, err := c.ReadGroups(context.Background(), &model.GroupsFilter{Name: optionalString("group-1-2-3")})

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
		groups, err := c.ReadGroups(context.Background(), &model.GroupsFilter{Name: optionalString(groupName)})

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf("failed to read group with name %s: query result is empty", groupName))
	})
}

func TestClientGroupsReadByNameRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups By Name - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		const groupName = "group-name"
		groups, err := c.ReadGroups(context.Background(), &model.GroupsFilter{Name: optionalString(groupName)})

		assert.Nil(t, groups)
		assert.EqualError(t, err, graphqlErr(c, "failed to read group with name "+groupName, errBadRequest))
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
			httpmock.NewErrorResponder(errBadRequest))

		groups, err := c.ReadGroups(context.Background(), nil)

		assert.Nil(t, groups)
		assert.EqualError(t, err, graphqlErr(c, "failed to read group with id All", errBadRequest))
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

		groups, err := c.ReadGroups(context.Background(), nil)

		assert.EqualError(t, err, fmt.Sprintf("failed to read group with id All: %v", client.ErrGraphqlResultIsEmpty))
		assert.Nil(t, groups)
	})
}

func TestClientDeleteGroupUsers(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Users", func(t *testing.T) {
		jsonResponse := `{
          "data": {
            "groupUpdate": {
              "ok": true,
              "error": null,
              "entity": {
                "id": "group-1",
                "name": "group-1"
              }
            }
          }
        }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse),
		)

		err := c.DeleteGroupUsers(context.Background(), "group-1", []string{"user-1"})

		assert.NoError(t, err)
	})
}

func TestClientDeleteGroupUsersEmptyUsers(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Users - Empty Users", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.DeleteGroupUsers(context.Background(), "group-1", nil)

		assert.NoError(t, err)
	})
}

func TestClientDeleteGroupUsersEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Users - Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.DeleteGroupUsers(context.Background(), "", []string{"user-1"})

		assert.EqualError(t, err, "failed to update group: id is empty")
	})
}

func TestClientDeleteGroupUsersRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Users - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest),
		)

		err := c.DeleteGroupUsers(context.Background(), "group-1", []string{"user-1"})

		assert.EqualError(t, err, graphqlErr(c, "failed to update group with id group-1", errBadRequest))
	})
}

func TestClientDeleteGroupUsersResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Users - Response Error", func(t *testing.T) {
		jsonResponse := `{
          "data": {
            "groupUpdate": {
              "ok": false,
              "error": "bad error",
              "entity": null
            }
          }
        }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse),
		)

		err := c.DeleteGroupUsers(context.Background(), "group-1", []string{"user-1"})

		assert.EqualError(t, err, `failed to update group with id group-1: bad error`)
	})
}

func TestClientDeleteGroupUsersEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Group Users - Empty Response", func(t *testing.T) {
		jsonResponse := `{
          "data": {
            "groupUpdate": {
              "ok": true,
              "error": null,
              "entity": null
            }
          }
        }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse),
		)

		err := c.DeleteGroupUsers(context.Background(), "group-1", []string{"user-1"})

		assert.EqualError(t, err, `failed to update group with id group-1: query result is empty`)
	})
}
