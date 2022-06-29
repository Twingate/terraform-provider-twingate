package twingate

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func newTestResource() *Resource {
	protocols := newProtocolsInput()
	protocols.TCP.Policy = policyAllowAll
	protocols.UDP.Policy = policyAllowAll

	groups := make([]*graphql.ID, 0)
	group := graphql.ID(b64.StdEncoding.EncodeToString([]byte("testgroup")))
	groups = append(groups, &group)

	return &Resource{
		ID:              graphql.ID("test"),
		RemoteNetworkID: graphql.ID("test"),
		Address:         "test",
		Name:            "testName",
		GroupsIds:       groups,
		Protocols:       protocols,
		IsActive:        true,
	}
}

func TestConvertToGraphqlUDPError(t *testing.T) {
	t.Run("Test Twingate Resource : Convert to GraphQL UDP Error", func(t *testing.T) {
		spi := &StringProtocolsInput{
			UDPPolicy: ".......",
			UDPPorts:  []string{"test-me"},
		}

		pi, err := spi.convertToGraphql()
		assert.Nil(t, pi)
		assert.Error(t, err)
	})
}

func TestConvertToGraphqlTCPError(t *testing.T) {
	t.Run("Test Twingate Resource : Convert to GraphQL TCP Error", func(t *testing.T) {
		spi := &StringProtocolsInput{
			TCPPolicy: ".......",
			TCPPorts:  []string{"test-me"},
		}

		pi, err := spi.convertToGraphql()
		assert.Nil(t, pi)
		assert.Error(t, err)
	})
}

func TestParsePortsToGraphql(t *testing.T) {
	t.Run("Test Twingate Resource : Parse Ports to GraphQL ", func(t *testing.T) {
		pri := []*PortRangeInput{}

		single := &PortRangeInput{
			Start: graphql.Int(80),
			End:   graphql.Int(80),
		}

		multi := &PortRangeInput{
			Start: graphql.Int(81),
			End:   graphql.Int(82),
		}

		pri = append(pri, single)
		pri = append(pri, multi)

		emptyPorts, err := convertPorts(make([]string, 0))
		assert.NoError(t, err)
		assert.Len(t, emptyPorts, 0)
		vars := []string{"80", "81-82"}
		ports, err := convertPorts(vars)
		assert.Equal(t, ports, pri)
		assert.NoError(t, err)
	})
}

func TestParseErrorPortsToGraphql(t *testing.T) {
	errString := func(portRange, port string) string {
		return fmt.Sprintf(`failed to parse protocols port range "%s": port is not a valid integer: strconv.ParseInt: parsing "%s": invalid syntax`, portRange, port)
	}

	t.Run("Test Twingate Resource : Client Resource Parse Ports to GraphQL Error", func(t *testing.T) {
		vars := []string{"foo"}
		_, err := convertPorts(vars)
		assert.EqualError(t, err, errString("foo", "foo"))

		vars = []string{"10-9"}
		_, err = convertPorts(vars)
		assert.EqualError(t, err, "failed to parse protocols port range \"10-9\": ports 10, 9 needs to be in a rising sequence")

		vars = []string{"abc-12345"}
		_, err = convertPorts(vars)
		assert.EqualError(t, err, errString("abc-12345", "abc"))

		vars = []string{"12345-abc"}
		_, err = convertPorts(vars)
		assert.EqualError(t, err, errString("12345-abc", "abc"))

		vars = []string{"1-999999"}
		_, err = convertPorts(vars)
		assert.EqualError(t, err, "failed to parse protocols port range \"1-999999\": port 999999 not in the range of 0-65535")

	})
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
		resource := newTestResource()

		err := client.createResource(context.Background(), resource)

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
		resource := newTestResource()

		err := client.createResource(context.Background(), resource)

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
		resource := newTestResource()

		err := client.createResource(context.Background(), resource)

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
			"type": "DNS",
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
	}`, policyAllowAll, policyRestricted)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceOkJson))

		resource, err := client.readResource(context.Background(), "resource1")
		tcpPorts, _ := resource.Protocols.TCP.buildPortsRange()
		assert.Nil(t, err)
		assert.EqualValues(t, graphql.ID("resource1"), resource.ID)
		assert.Contains(t, resource.stringGroups(), "group1")
		assert.Contains(t, tcpPorts, "8080-8090")
		assert.EqualValues(t, resource.Address, "test.com")
		assert.EqualValues(t, resource.RemoteNetworkID, graphql.ID("network1"))
		assert.Len(t, resource.Protocols.UDP.Ports, 0)
		assert.EqualValues(t, resource.Name, "test resource")
	})
}

func TestClientResourceReadTooManyGroups(t *testing.T) {
	t.Run("Test Twingate Resource : Read To Many Groups", func(t *testing.T) {
		// response JSON
		createResourceOkJson := fmt.Sprintf(`{
	  "data": {
		"resource": {
		  "id": "resource1",
		  "name": "test resource",
		  "address": {
			"type": "DNS",
			"value": "test.com"
		  },
		  "remoteNetwork": {
			"id": "network1"
		  },
		  "groups": {
			"pageInfo": {
			  "hasNextPage": true
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
	}`, policyAllowAll, policyRestricted)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceOkJson))
		resourceID := "resource1"

		resource, err := client.readResource(context.Background(), resourceID)
		assert.Nil(t, resource)
		assert.EqualError(t, err, fmt.Sprintf("failed to read resource with id %s: provider does not support more than 50 groups per resource", resourceID))
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
		resourceID := "resource1"

		resource, err := client.readResource(context.Background(), resourceID)

		assert.Nil(t, resource)
		assert.EqualError(t, err, ErrGraphqlResourceNotFound.Error())
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

		resource, err := client.readResource(context.Background(), "")

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

		resource, err := client.readResource(context.Background(), resourceID)

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
				"error" : null
			}
		}
	}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceUpdateOkJson))
		resource := newTestResource()

		err := client.updateResource(context.Background(), resource)

		assert.Nil(t, err)
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
		resource := newTestResource()

		err := client.updateResource(context.Background(), resource)

		assert.EqualError(t, err, "failed to update resource with id test: cant update resource")
	})
}

func TestClientResourceUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))
		resource := newTestResource()

		err := client.updateResource(context.Background(), resource)

		assert.EqualError(t, err, fmt.Sprintf(`failed to update resource with id %v: Post "%s": error_1`, resource.ID, client.GraphqlServerURL))
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

		err := client.deleteResource(context.Background(), "resource1")

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

		err := client.deleteResource(context.Background(), resourceID)

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

		err := client.deleteResource(context.Background(), resourceID)

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

		err := client.deleteResource(context.Background(), "")

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

		edges, err := client.readResources(context.Background())
		assert.NoError(t, err)

		mockMap := make(map[graphql.ID]graphql.String)
		mockMap["resource1"] = "tf-acc-resource1"
		mockMap["resource2"] = "resource2"
		mockMap["resource3"] = "tf-acc-resource3"

		for _, elem := range edges {
			name := mockMap[elem.Node.ID]
			assert.Equal(t, name, elem.Node.Name)
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

		resources, err := client.readResources(context.Background())

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

		err := client.updateResourceActiveState(context.Background(), resource)

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

		err := client.updateResourceActiveState(context.Background(), resource)

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

		err := client.updateResourceActiveState(context.Background(), resource)

		assert.EqualError(t, err, fmt.Sprintf(`failed to update resource with id %v: Post "%s": error_1`, resource.ID, client.GraphqlServerURL))
	})
}
