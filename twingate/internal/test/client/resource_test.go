package client

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func newTestResource() *model.Resource {
	groups := []string{b64.StdEncoding.EncodeToString([]byte("testgroup"))}

	return &model.Resource{
		ID:              "test",
		RemoteNetworkID: "test",
		Address:         "test",
		Name:            "testName",
		Groups:          groups,
		Protocols:       model.DefaultProtocols(),
		IsActive:        true,
	}
}

func TestClientResourceCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Create Ok", func(t *testing.T) {
		// response JSON
		createResourceOkJson := `{
		  "data": {
		    "resourceCreate": {
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
			httpmock.NewStringResponder(200, createResourceOkJson))

		resource, err := client.CreateResource(context.Background(), &model.Resource{ID: "test-id"})

		assert.Nil(t, err)
		assert.EqualValues(t, "test-id", resource.ID)
	})
}

func TestClientResourceCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Create Error", func(t *testing.T) {
		// response JSON
		createResourceErrorJson := `{
		  "data": {
		    "resourceCreate": {
		      "entity": {
		        "id": "test-id"
		      },
		      "ok": false,
		      "error": "something went wrong"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceErrorJson))

		_, err := client.CreateResource(context.Background(), &model.Resource{ID: "test-id"})

		assert.EqualError(t, err, "failed to create resource: something went wrong")
	})
}

func TestClientResourceCreateEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Create - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resourceCreate": {
		      "entity": null,
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		_, err := client.CreateResource(context.Background(), &model.Resource{ID: "test-id"})

		assert.EqualError(t, err, "failed to create resource: query result is empty")
	})
}

func TestClientResourceCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Create Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		_, err := client.CreateResource(context.Background(), &model.Resource{ID: "test-id"})

		assert.EqualError(t, err, graphqlErr(client, "failed to create resource", errBadRequest))
	})
}

func TestClientResourceReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Client Resource Ok", func(t *testing.T) {
		var defaultBool bool

		expected := &model.Resource{
			ID:              "resource1",
			Name:            "test resource",
			Address:         "test.com",
			RemoteNetworkID: "network1",
			Groups: []string{
				"group1", "group2",
			},
			Protocols: &model.Protocols{
				UDP: &model.Protocol{
					Ports:  []*model.PortRange{},
					Policy: model.PolicyAllowAll,
				},
				TCP: &model.Protocol{
					Ports: []*model.PortRange{
						{Start: 80, End: 80},
						{Start: 8080, End: 8090},
					},
					Policy: model.PolicyRestricted,
				},
				AllowIcmp: true,
			},
			IsVisible:                &defaultBool,
			IsBrowserShortcutEnabled: &defaultBool,
		}

		jsonResponse := fmt.Sprintf(`{
		  "data": {
		    "resource": {
		      "id": "resource1",
		      "name": "test resource",
		      "address": {
		        "value": "test.com"
		      },
		      "remoteNetwork": {
		        "id": "network1"
		      },
		      "access": {
		        "pageInfo": {
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "__typename": "Group",
		              "id": "group1"
		            }
		          },
		          {
		            "node": {
		              "__typename": "Group",
		              "id": "group2"
		            }
		          }
		        ]
		      },
		      "protocols": {
		        "udp": {
		          "ports": [],
		          "policy": "%s"
		        },
		        "tcp": {
		          "ports": [
		            {
		              "end": 80,
		              "start": 80
		            },
		            {
		              "end": 8090,
		              "start": 8080
		            }
		          ],
		          "policy": "%s"
		        },
		        "allowIcmp": true
		      }
		    }
		  }
		}`, model.PolicyAllowAll, model.PolicyRestricted)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		resource, err := client.ReadResource(context.Background(), "resource1")
		assert.Nil(t, err)
		assert.EqualValues(t, expected, resource)
	})
}

func TestClientResourceReadAllGroups(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Groups", func(t *testing.T) {
		var defaultBool bool

		expected := &model.Resource{
			ID:              "resource1",
			Name:            "test resource",
			Address:         "test.com",
			RemoteNetworkID: "network1",
			Groups: []string{
				"group1", "group2", "group3", "group4",
			},
			IsActive: true,
			Protocols: &model.Protocols{
				UDP: &model.Protocol{
					Ports:  []*model.PortRange{},
					Policy: model.PolicyAllowAll,
				},
				TCP: &model.Protocol{
					Ports: []*model.PortRange{
						{Start: 80, End: 80},
						{Start: 8080, End: 8090},
					},
					Policy: model.PolicyRestricted,
				},
				AllowIcmp: true,
			},
			IsVisible:                &defaultBool,
			IsBrowserShortcutEnabled: &defaultBool,
		}

		jsonResponse := fmt.Sprintf(`{
		  "data": {
		    "resource": {
		      "id": "resource1",
		      "name": "test resource",
		      "address": {
		        "value": "test.com"
		      },
		      "remoteNetwork": {
		        "id": "network1"
		      },
		      "access": {
		        "pageInfo": {
		          "endCursor": "cur001",
		          "hasNextPage": true
		        },
		        "edges": [
		          {
		            "node": {
		              "__typename": "Group",
		              "id": "group1"
		            }
		          },
		          {
		            "node": {
		              "__typename": "Group",
		              "id": "group2"
		            }
		          }
		        ]
		      },
		      "isActive": true,
		      "protocols": {
		        "udp": {
		          "ports": [],
		          "policy": "%s"
		        },
		        "tcp": {
		          "ports": [
		            {
		              "end": 80,
		              "start": 80
		            },
		            {
		              "end": 8090,
		              "start": 8080
		            }
		          ],
		          "policy": "%s"
		        },
		        "allowIcmp": true
		      }
		    }
		  }
		}`, model.PolicyAllowAll, model.PolicyRestricted)

		nextPageJson := fmt.Sprintf(`{
		  "data": {
		    "resource": {
		      "id": "resource1",
		      "access": {
		        "pageInfo": {
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "__typename": "Group",
		              "id": "group3"
		            }
		          },
		          {
		            "node": {
		              "__typename": "Group",
		              "id": "group4"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.ResponderFromMultipleResponses(
				[]*http.Response{
					httpmock.NewStringResponse(200, jsonResponse),
					httpmock.NewStringResponse(200, nextPageJson),
				}),
		)

		resource, err := client.ReadResource(context.Background(), "resource1")
		assert.NoError(t, err)
		assert.Equal(t, expected, resource)
	})
}

func TestClientResourceReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Read Error", func(t *testing.T) {
		// response JSON
		createResourceErrorJson := `{
		  "data": {
		    "resource": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceErrorJson))
		const resourceID = "resource1"

		resource, err := client.ReadResource(context.Background(), resourceID)

		assert.Nil(t, resource)
		assert.EqualError(t, err, fmt.Sprintf("failed to read resource with id %s: query result is empty", resourceID))
	})
}

func TestClientResourceEmptyReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Empty Read Error", func(t *testing.T) {
		// response JSON
		createResourceErrorJson := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceErrorJson))

		resource, err := client.ReadResource(context.Background(), "")

		assert.Nil(t, resource)
		assert.EqualError(t, err, "failed to read resource: id is empty")
	})
}

func TestClientResourceReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Read Request Error", func(t *testing.T) {
		// response JSON
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))
		resourceID := "test-id"

		resource, err := client.ReadResource(context.Background(), resourceID)

		assert.Nil(t, resource)
		assert.EqualError(t, err, graphqlErr(client, "failed to read resource with id "+resourceID, errBadRequest))
	})
}

func TestClientResourceReadGroupsRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups - Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resource": {
		      "id": "resource1",
		      "access": {
		        "pageInfo": {
		          "endCursor": "cur001",
		          "hasNextPage": true
		        },
		        "edges": []
		      }
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		resource, err := client.ReadResource(context.Background(), "resource1")
		assert.Nil(t, resource)
		assert.EqualError(t, err, graphqlErr(client, "failed to read resource with id resource1", errBadRequest))
	})
}

func TestClientResourceReadGroupsEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Read Groups - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resource": {
		      "id": "resource1",
		      "access": {
		        "pageInfo": {
		          "endCursor": "cur001",
		          "hasNextPage": true
		        },
		        "edges": []
		      }
		    }
		  }
		}`

		nextPage := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewStringResponder(200, nextPage),
			),
		)

		resource, err := client.ReadResource(context.Background(), "resource1")
		assert.Nil(t, resource)
		assert.EqualError(t, err, "failed to read resource with id resource1: query result is empty")
	})
}

func TestClientResourceUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Ok", func(t *testing.T) {
		// response JSON
		createResourceUpdateOkJson := `{
		  "data": {
		    "resourceUpdate": {
		      "ok" : true,
		      "error" : null,
		      "entity": {}
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceUpdateOkJson))

		_, err := client.UpdateResource(context.Background(), newTestResource())
		assert.NoError(t, err)
	})
}

func TestClientResourceUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Error", func(t *testing.T) {
		// response JSON
		createResourceUpdateErrorJson := `{
		  "data": {
		    "resourceUpdate": {
		      "ok" : false,
		      "error" : "cant update resource"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceUpdateErrorJson))

		_, err := client.UpdateResource(context.Background(), newTestResource())

		assert.EqualError(t, err, "failed to update resource with id test: cant update resource")
	})
}

func TestClientResourceUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		req := newTestResource()
		_, err := client.UpdateResource(context.Background(), req)

		assert.EqualError(t, err, graphqlErr(client, "failed to update resource with id "+req.ID, errBadRequest))
	})
}

func TestClientResourceUpdateEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update - Empty Response", func(t *testing.T) {
		emptyResponse := `{
		  "data": {
		    "resourceUpdate": {
		      "ok": true,
		      "entity": null
		    }
		  }
		}`
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, emptyResponse))

		_, err := client.UpdateResource(context.Background(), newTestResource())

		assert.EqualError(t, err, "failed to update resource with id test: query result is empty")
	})
}

func TestClientResourceUpdateFetchGroupsError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update - Fetch Groups Error", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "resourceUpdate": {
		      "ok" : true,
		      "error" : null,
		      "entity": {
		        "id": "test",
		        "access": {
		          "pageInfo": {
		            "endCursor": "cur001",
		            "hasNextPage": true
		          },
		          "edges": [
		            {
		              "node": {
		                "__typename": "Group",
		                "id": "group-1"
		              }
		            }
		          ]
		        }
		      }
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewErrorResponder(errBadRequest),
			))

		_, err := client.UpdateResource(context.Background(), newTestResource())

		assert.EqualError(t, err, graphqlErr(client, "failed to read resource with id test", errBadRequest))
	})
}

func TestClientResourceDeleteOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Delete Ok", func(t *testing.T) {
		// response JSON
		createResourceDeleteOkJson := `{
		  "data": {
		    "resourceDelete": {
		      "ok" : true,
		      "error" : null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceDeleteOkJson))

		err := client.DeleteResource(context.Background(), "resource1")

		assert.Nil(t, err)
	})
}

func TestClientResourceDeleteError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Delete Error", func(t *testing.T) {
		// response JSON
		createResourceDeleteErrorJson := `{
		  "data": {
		    "resourceDelete": {
		      "ok" : false,
		      "error" : "cant delete resource"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceDeleteErrorJson))

		resourceID := "resource1"
		err := client.DeleteResource(context.Background(), resourceID)

		assert.EqualError(t, err, fmt.Sprintf("failed to delete resource with id %s: cant delete resource", resourceID))
	})
}

func TestClientResourceDeleteRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Delete Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		resourceID := "resource1"
		err := client.DeleteResource(context.Background(), resourceID)

		assert.EqualError(t, err, graphqlErr(client, "failed to delete resource with id "+resourceID, errBadRequest))
	})
}

func TestClientResourceEmptyDeleteError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Empty Delete Error", func(t *testing.T) {
		// response JSON
		createResourceDeleteErrorJson := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceDeleteErrorJson))

		err := client.DeleteResource(context.Background(), "")

		assert.EqualError(t, err, "failed to delete resource: id is empty")
	})
}

func TestClientResourcesReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Read All Ok", func(t *testing.T) {
		var defaultBool bool

		expected := []*model.Resource{
			{ID: "resource1", Name: "tf-acc-resource1", IsVisible: &defaultBool, IsBrowserShortcutEnabled: &defaultBool, Protocols: model.DefaultProtocols()},
			{ID: "resource2", Name: "resource2", IsVisible: &defaultBool, IsBrowserShortcutEnabled: &defaultBool, Protocols: model.DefaultProtocols()},
			{ID: "resource3", Name: "tf-acc-resource3", IsVisible: &defaultBool, IsBrowserShortcutEnabled: &defaultBool, Protocols: model.DefaultProtocols()},
			{ID: "resource4", Name: "tf-acc-resource4", IsVisible: &defaultBool, IsBrowserShortcutEnabled: &defaultBool, Protocols: model.DefaultProtocols()},
			{ID: "resource5", Name: "tf-acc-resource5", IsVisible: &defaultBool, IsBrowserShortcutEnabled: &defaultBool, Protocols: model.DefaultProtocols()},
		}

		// response JSON
		readResourcesOkJson := `{
		  "data": {
		    "resources": {
		      "pageInfo": {
		        "endCursor": "cur001",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "resource1",
		            "name": "tf-acc-resource1"
		          }
		        },
		        {
		          "node": {
		            "id": "resource2",
		            "name": "resource2"
		          }
		        },
		        {
		          "node": {
		            "id": "resource3",
		            "name": "tf-acc-resource3"
		          }
		        }
		      ]
		    }
		  }
		}`

		nextPage := `{
		  "data": {
		    "resources": {
		      "pageInfo": {
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "resource4",
		            "name": "tf-acc-resource4"
		          }
		        },
		        {
		          "node": {
		            "id": "resource5",
		            "name": "tf-acc-resource5"
		          }
		        }
		      ]
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.ResponderFromMultipleResponses(
				[]*http.Response{
					httpmock.NewStringResponse(200, readResourcesOkJson),
					httpmock.NewStringResponse(200, nextPage),
				}),
		)

		resources, err := client.ReadResources(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, resources)
	})
}

func TestClientResourcesReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resources Read Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		resources, err := client.ReadResources(context.Background())

		assert.Nil(t, resources)
		assert.EqualError(t, err, graphqlErr(client, "failed to read resource with id All", errBadRequest))
	})
}

func TestClientResourcesReadAllRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Read All - Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resources": {
		      "pageInfo": {
		        "endCursor": "cur001",
		        "hasNextPage": true
		      },
		      "edges": []
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		resources, err := client.ReadResources(context.Background())

		assert.Nil(t, resources)
		assert.EqualError(t, err, graphqlErr(client, "failed to read resource", errBadRequest))
	})
}

func TestClientResourcesReadAllEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Read All - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resources": {
		      "pageInfo": {
		        "endCursor": "cur001",
		        "hasNextPage": true
		      },
		      "edges": []
		    }
		  }
		}`

		emptyResponse := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewStringResponder(200, emptyResponse),
			),
		)

		resources, err := client.ReadResources(context.Background())

		assert.Nil(t, resources)
		assert.EqualError(t, err, `failed to read resource: query result is empty`)
	})
}

func TestClientResourceUpdateActiveStateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Active State Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resourceUpdate": {
		      "ok" : true,
		      "error" : null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))
		resource := newTestResource()

		err := client.UpdateResourceActiveState(context.Background(), resource)

		assert.Nil(t, err)
	})
}

func TestClientResourceUpdateActiveStateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Active State Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resourceUpdate": {
		      "ok" : false,
		      "error" : "cant update resource"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))
		resource := newTestResource()

		err := client.UpdateResourceActiveState(context.Background(), resource)

		assert.EqualError(t, err, "failed to update resource with id test: cant update resource")
	})
}

func TestClientResourceUpdateActiveStateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Active State Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))
		resource := newTestResource()

		err := client.UpdateResourceActiveState(context.Background(), resource)

		assert.EqualError(t, err, graphqlErr(client, "failed to update resource with id "+resource.ID, errBadRequest))
	})
}

// Read resource without groups

func TestClientResourceReadWithoutGroupsError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Read Without Groups Error", func(t *testing.T) {
		responseJSON := `{
		  "data": {
		    "resource": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, responseJSON))
		resourceID := "resource1"

		resource, err := client.ReadResource(context.Background(), resourceID)

		assert.Nil(t, resource)
		assert.EqualError(t, err, fmt.Sprintf("failed to read resource with id %s: query result is empty", resourceID))
	})
}

func TestClientResourceEmptyReadWithoutGroupsError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Empty Read Without Groups Error", func(t *testing.T) {
		responseJSON := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, responseJSON))

		resource, err := client.ReadResource(context.Background(), "")

		assert.Nil(t, resource)
		assert.EqualError(t, err, "failed to read resource: id is empty")
	})
}

func TestClientResourceReadWithoutGroupsRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Read Without Groups Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))
		resourceID := "test-id"

		resource, err := client.ReadResource(context.Background(), resourceID)

		assert.Nil(t, resource)
		assert.EqualError(t, err, graphqlErr(client, "failed to read resource with id "+resourceID, errBadRequest))
	})
}

func TestClientResourceReadWithoutGroupsOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Client Resource Resource Without Groups Ok", func(t *testing.T) {
		responseJSON := fmt.Sprintf(`{
		  "data": {
		    "resource": {
		      "id": "resource1",
		      "name": "test resource",
		      "address": {
		        "value": "test.com"
		      },
		      "remoteNetwork": {
		        "id": "network1"
		      },
		      "protocols": {
		        "udp": {
		          "ports": [],
		          "policy": "%s"
		        },
		        "tcp": {
		          "ports": [
		            {
		              "end": 80,
		              "start": 80
		            },
		            {
		              "end": 8090,
		              "start": 8080
		            }
		          ],
		          "policy": "%s"
		        },
		        "allowIcmp": true
		      }
		    }
		  }
		}
	  }
	}`, model.PolicyAllowAll, model.PolicyRestricted)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, responseJSON))

		resource, err := client.ReadResource(context.Background(), "resource1")
		assert.Nil(t, err)
		assert.EqualValues(t, "resource1", resource.ID)
		assert.EqualValues(t, []string{"80", "8080-8090"}, resource.Protocols.TCP.PortsToString())
		assert.EqualValues(t, resource.Address, "test.com")
		assert.EqualValues(t, resource.RemoteNetworkID, graphql.ID("network1"))
		assert.Len(t, resource.Protocols.UDP.Ports, 0)
		assert.EqualValues(t, resource.Name, "test resource")
	})
}

func TestClientResourcesReadByNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Resources By Name - Ok", func(t *testing.T) {
		var defaultBool bool

		expected := []*model.Resource{
			{
				ID: "id-1", Name: "resource-test", Address: "internal.int",
				Protocols: &model.Protocols{
					TCP: &model.Protocol{
						Policy: model.PolicyRestricted,
						Ports: []*model.PortRange{
							{Start: 80, End: 80},
							{Start: 82, End: 83},
						},
					},
					UDP: &model.Protocol{
						Policy: model.PolicyAllowAll,
						Ports:  []*model.PortRange{},
					},
				},
				RemoteNetworkID:          "UmVtb3RlTmV0d29yazo0MDEzOQ==",
				IsVisible:                &defaultBool,
				IsBrowserShortcutEnabled: &defaultBool,
			},
			{
				ID: "id-2", Name: "resource-test", Address: "internal.int",
				Protocols: &model.Protocols{
					TCP: &model.Protocol{
						Policy: model.PolicyRestricted,
						Ports: []*model.PortRange{
							{Start: 80, End: 80},
							{Start: 82, End: 83},
						},
					},
					UDP: &model.Protocol{
						Policy: model.PolicyAllowAll,
						Ports:  []*model.PortRange{},
					},
				},
				RemoteNetworkID:          "UmVtb3RlTmV0d29yazo0MDEzOQ==",
				IsVisible:                &defaultBool,
				IsBrowserShortcutEnabled: &defaultBool,
			},
			{
				ID: "id-3", Name: "resource-test", Address: "internal.int",
				Protocols: &model.Protocols{
					TCP: &model.Protocol{
						Policy: model.PolicyRestricted,
						Ports: []*model.PortRange{
							{Start: 80, End: 80},
							{Start: 82, End: 83},
						},
					},
					UDP: &model.Protocol{
						Policy: model.PolicyAllowAll,
						Ports:  []*model.PortRange{},
					},
				},
				RemoteNetworkID:          "UmVtb3RlTmV0d29yazo0MDEzOQ==",
				IsVisible:                &defaultBool,
				IsBrowserShortcutEnabled: &defaultBool,
			},
		}

		jsonResponse := `{
		  "data": {
		    "resources": {
		      "pageInfo": {
		        "endCursor": "cur-01",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "resource-test",
		            "address": {
		              "value": "internal.int"
		            },
		            "protocols": {
		              "tcp": {
		                "policy": "RESTRICTED",
		                "ports": [
		                  {
		                    "start": 80,
		                    "end": 80
		                  },
		                  {
		                    "start": 82,
		                    "end": 83
		                  }
		                ]
		              },
		              "udp": {
		                "policy": "ALLOW_ALL",
		                "ports": []
		              }
		            },
		            "remoteNetwork": {
		              "id": "UmVtb3RlTmV0d29yazo0MDEzOQ=="
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "resource-test",
		            "address": {
		              "value": "internal.int"
		            },
		            "protocols": {
		              "tcp": {
		                "policy": "RESTRICTED",
		                "ports": [
		                  {
		                    "start": 80,
		                    "end": 80
		                  },
		                  {
		                    "start": 82,
		                    "end": 83
		                  }
		                ]
		              },
		              "udp": {
		                "policy": "ALLOW_ALL",
		                "ports": []
		              }
		            },
		            "remoteNetwork": {
		              "id": "UmVtb3RlTmV0d29yazo0MDEzOQ=="
		            }
		          }
		        }
		      ]
		    }
		  }
		}`

		nextPage := `{
		  "data": {
			"resources": {
			  "pageInfo": {
				"hasNextPage": false
			  },
			  "edges": [
				{
				  "node": {
					"id": "id-3",
					"name": "resource-test",
					"address": {
					  "value": "internal.int"
					},
					"protocols": {
					  "tcp": {
						"policy": "RESTRICTED",
						"ports": [
						  {
							"start": 80,
							"end": 80
						  },
						  {
							"start": 82,
							"end": 83
						  }
						]
					  },
					  "udp": {
						"policy": "ALLOW_ALL",
						"ports": []
					  }
					},
					"remoteNetwork": {
					  "id": "UmVtb3RlTmV0d29yazo0MDEzOQ=="
					}
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

		resources, err := client.ReadResourcesByName(context.Background(), "resource-test", "")

		assert.Nil(t, err)
		assert.Equal(t, expected, resources)
	})
}

func TestClientResourcesReadByNameEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Read Resources By Name - Empty Result", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resources": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		resources, err := client.ReadResourcesByName(context.Background(), "resource-name", "")

		assert.Nil(t, resources)
		assert.EqualError(t, err, "failed to read resource with id All: query result is empty")
	})
}

func TestClientResourcesReadByNameRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Resources By Name - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		groups, err := client.ReadResourcesByName(context.Background(), "resource-name", "")

		assert.Nil(t, groups)
		assert.EqualError(t, err, graphqlErr(client, "failed to read resource with id All", errBadRequest))
	})
}

func TestClientResourcesReadByNameErrorEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource : Read Resources By Name - Error Empty Name", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resources": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		groups, err := client.ReadResourcesByName(context.Background(), "", "")

		assert.Nil(t, groups)
		assert.EqualError(t, err, "failed to read resource with id All: query result is empty")
	})
}

func TestClientResourcesReadByNameAllEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Read Resources By Name All - Empty Result", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resources": {
		      "pageInfo": {
		        "endCursor": "cur-01",
		        "hasNextPage": true
		      },
		      "edges": [{}]
		    }
		  }
		}`

		emptyResponse := `{
		  "data": {
		    "resources": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewStringResponder(200, emptyResponse),
			),
		)

		resources, err := client.ReadResourcesByName(context.Background(), "resource-name", "")

		assert.Nil(t, resources)
		assert.EqualError(t, err, "failed to read resource with id All: query result is empty")
	})
}

func TestClientResourcesReadByNameAllRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Resources By Name All - Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resources": {
		      "pageInfo": {
		        "endCursor": "cur-01",
		        "hasNextPage": true
		      },
		      "edges": [{}]
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		resources, err := client.ReadResourcesByName(context.Background(), "resource-name", "")

		assert.Nil(t, resources)
		assert.EqualError(t, err, graphqlErr(client, "failed to read resource with id All", errBadRequest))
	})
}

func TestClientRemoveResourceAccessOk(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Resource Access - Ok", func(t *testing.T) {
		response := `{
		  "data": {
		    "resourceAccessRemove": {
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, response),
		)

		err := client.RemoveResourceAccess(context.Background(),
			"resource-1",
			[]string{"id-1", "id-2"},
		)

		assert.NoError(t, err)
	})
}

func TestClientRemoveResourceAccessWithEmptyList(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Resource Access - With Empty List", func(t *testing.T) {
		client := newHTTPMockClient()

		err := client.RemoveResourceAccess(context.Background(),
			"resource-1",
			[]string{},
		)

		assert.NoError(t, err)
	})
}

func TestClientRemoveResourceAccessWithoutID(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Resource Access - Without ID", func(t *testing.T) {
		client := newHTTPMockClient()

		err := client.RemoveResourceAccess(context.Background(),
			"",
			[]string{"id-1"},
		)

		assert.Error(t, err, "failed to delete resource access: id is empty")
	})
}

func TestClientRemoveResourceAccessRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Resource Access - Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := client.RemoveResourceAccess(context.Background(),
			"resource-1",
			[]string{"id-1", "id-2"},
		)

		assert.EqualError(t, err, graphqlErr(client, "failed to delete resource access with id resource-1", errBadRequest))
	})
}

func TestClientRemoveResourceAccessResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Remove Resource Access - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resourceAccessRemove": {
		      "ok": false,
		      "error": "response error"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := client.RemoveResourceAccess(context.Background(),
			"resource-1",
			[]string{"id-1", "id-2"},
		)
		assert.EqualError(t, err, `failed to delete resource access with id resource-1: response error`)
	})
}

func TestClientAddResourceAccessOk(t *testing.T) {
	t.Run("Test Twingate Resource : Add Resource Access - Ok", func(t *testing.T) {
		response := `{
		  "data": {
		    "resourceAccessAdd": {
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, response),
		)

		err := client.AddResourceAccess(context.Background(),
			"resource-1",
			[]string{"id-1", "id-2"},
		)

		assert.NoError(t, err)
	})
}

func TestClientAddResourceAccessWithEmptyList(t *testing.T) {
	t.Run("Test Twingate Resource : Add Resource Access - With Empty List", func(t *testing.T) {
		client := newHTTPMockClient()

		err := client.AddResourceAccess(context.Background(),
			"resource-1",
			[]string{},
		)

		assert.NoError(t, err)
	})
}

func TestClientAddResourceAccessWithoutID(t *testing.T) {
	t.Run("Test Twingate Resource : Add Resource Access - Without ID", func(t *testing.T) {
		client := newHTTPMockClient()

		err := client.AddResourceAccess(context.Background(),
			"",
			[]string{"id-1"},
		)

		assert.Error(t, err, "failed to update resource access: id is empty")
	})
}

func TestClientAddResourceAccessRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Add Resource Access - Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := client.AddResourceAccess(context.Background(),
			"resource-1",
			[]string{"id-1", "id-2"},
		)

		assert.EqualError(t, err, graphqlErr(client, "failed to update resource access with id resource-1", errBadRequest))
	})
}

func TestClientAddResourceAccessResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Add Resource Access - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "resourceAccessAdd": {
		      "ok": false,
		      "error": "response error"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := client.AddResourceAccess(context.Background(),
			"resource-1",
			[]string{"id-1", "id-2"},
		)
		assert.EqualError(t, err, `failed to update resource access with id resource-1: response error`)
	})
}
