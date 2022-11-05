package client

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
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

func TestClientResourceCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Create Request Error", func(t *testing.T) {
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
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, createResourceErrorJson)
				return resp, errors.New("error_1")
			})

		_, err := client.CreateResource(context.Background(), &model.Resource{ID: "test-id"})

		assert.EqualError(t, err, fmt.Sprintf(`failed to create resource: Post "%s": error_1`, client.GraphqlServerURL))
	})
}

func TestClientResourceReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Client Resource Ok", func(t *testing.T) {
		// response JSON
		createResourceOkJson := fmt.Sprintf(`{
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
		  "groups": {
			"pageInfo": {
			  "hasNextPage": false
			},
			"edges": [
			  {
				"node": {
				  "id": "group1"
				}
			  },
			  {
				"node": {
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
			httpmock.NewStringResponder(200, createResourceOkJson))

		resource, err := client.ReadResource(context.Background(), "resource1")
		assert.Nil(t, err)
		assert.EqualValues(t, "resource1", resource.ID)
		assert.EqualValues(t, []string{"group1", "group2"}, resource.Groups)
		assert.EqualValues(t, []string{"80", "8080-8090"}, resource.Protocols.TCP.PortsToString())
		assert.EqualValues(t, "test.com", resource.Address)
		assert.EqualValues(t, "network1", resource.RemoteNetworkID)
		assert.Len(t, resource.Protocols.UDP.Ports, 0)
		assert.EqualValues(t, resource.Name, "test resource")
	})
}

func TestClientResourceReadTooManyGroups(t *testing.T) {
	t.Run("Test Twingate Resource : Read To Many Groups", func(t *testing.T) {
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
		}

		// response JSON
		createResourceOkJson := fmt.Sprintf(`{
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
		  "groups": {
			"pageInfo": {
			  "endCursor": "cur001",
			  "hasNextPage": true
			},
			"edges": [
			  {
				"node": {
				  "id": "group1",
				  "name": "Group1 name"
				}
			  },
			  {
				"node": {
				  "id": "group2",
				  "name": "Group2 name"
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
	      "groups": {
	        "pageInfo": {
	          "hasNextPage": false
	        },
	        "edges": [
	          {
	            "node": {
	              "id": "group3",
	              "name": "Group3 name"
	            }
	          },
	          {
	            "node": {
	              "id": "group4",
	              "name": "Group4 name"
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
					httpmock.NewStringResponse(200, createResourceOkJson),
					httpmock.NewStringResponse(200, nextPageJson),
				},
				t.Log),
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
			httpmock.NewErrorResponder(errors.New("error_1")))
		resourceID := "test-id"

		resource, err := client.ReadResource(context.Background(), resourceID)

		assert.Nil(t, resource)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read resource with id %s: Post "%s": error_1`, resourceID, client.GraphqlServerURL))
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
			httpmock.NewErrorResponder(errors.New("error_1")))

		req := newTestResource()
		_, err := client.UpdateResource(context.Background(), req)

		assert.EqualError(t, err, fmt.Sprintf(`failed to update resource with id %v: Post "%s": error_1`, req.ID, client.GraphqlServerURL))
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
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, createResourceDeleteErrorJson)
				return resp, errors.New("error_1")
			})
		resourceID := "resource1"

		err := client.DeleteResource(context.Background(), resourceID)

		assert.EqualError(t, err, fmt.Sprintf(`failed to delete resource with id %s: Post "%s": error_1`, resourceID, client.GraphqlServerURL))
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
		// response JSON
		readResourcesOkJson := `{
	  "data": {
		"resources": {
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

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readResourcesOkJson))

		resources, err := client.ReadResources(context.Background())
		assert.NoError(t, err)

		mockMap := make(map[string]string)
		mockMap["resource1"] = "tf-acc-resource1"
		mockMap["resource2"] = "resource2"
		mockMap["resource3"] = "tf-acc-resource3"

		for _, elem := range resources {
			name := mockMap[elem.ID]
			assert.Equal(t, name, elem.Name)
		}
	})
}

func TestClientResourcesReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resources Read Request Error", func(t *testing.T) {
		// response JSON
		readResourcesOkJson := `{
	  "data": {
		"resources": {
		  "edges": [
			{
			  "node": {
				"id": "resource1",
				"name": "tf-acc-resource1"
			  }
			},
		  ]
		}
	  }
	}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, readResourcesOkJson)
				return resp, errors.New("error_1")
			})

		resources, err := client.ReadResources(context.Background())

		assert.Nil(t, resources)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read resource with id All: Post "%s": error_1`, client.GraphqlServerURL))
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
			httpmock.NewErrorResponder(errors.New("error_1")))
		resource := newTestResource()

		err := client.UpdateResourceActiveState(context.Background(), resource)

		assert.EqualError(t, err, fmt.Sprintf(`failed to update resource with id %v: Post "%s": error_1`, resource.ID, client.GraphqlServerURL))
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
			httpmock.NewErrorResponder(errors.New("error_1")))
		resourceID := "test-id"

		resource, err := client.ReadResource(context.Background(), resourceID)

		assert.Nil(t, resource)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read resource with id %s: Post "%s": error_1`, resourceID, client.GraphqlServerURL))
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
		const resourceName = "resource-1"
		ids := []string{"id-1", "id-2"}
		jsonResponse := fmt.Sprintf(`{
		  "data": {
			"resources": {
			  "edges": [
				{
				  "node": {
					"id": "%s",
					"name": "%s",
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
					"id": "%s",
					"name": "%s",
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
		}`, ids[0], resourceName, ids[1], resourceName)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		resources, err := client.ReadResourcesByName(context.Background(), resourceName)

		assert.Nil(t, err)
		assert.NotNil(t, resources)
		assert.Len(t, resources, len(ids))
		for i, id := range ids {
			assert.EqualValues(t, id, resources[i].ID)
			assert.EqualValues(t, resourceName, resources[i].Name)
		}
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

		resources, err := client.ReadResourcesByName(context.Background(), "resource-name")

		assert.Nil(t, resources)
		assert.EqualError(t, err, "failed to read resource with id All: query result is empty")
	})
}

func TestClientResourcesReadByNameRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Resources By Name - Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
			"resources": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, jsonResponse)
				return resp, errors.New("error_1")
			})

		groups, err := client.ReadResourcesByName(context.Background(), "resource-name")

		assert.Nil(t, groups)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read resource with id All: Post "%s": error_1`, client.GraphqlServerURL))
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

		groups, err := client.ReadResourcesByName(context.Background(), "")

		assert.Nil(t, groups)
		assert.EqualError(t, err, "failed to read resource with id All: query result is empty")
	})
}
