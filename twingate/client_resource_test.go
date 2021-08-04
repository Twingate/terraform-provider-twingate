package twingate

import (
	b64 "encoding/base64"
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func newTestResource() *Resource {
	protocols := newProcolsInput()
	protocols.TCP.Policy = "ALLOW_ALL"
	protocols.UDP.Policy = "ALLOW_ALL"

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
	t.Run("Test Twingate Resource : Client Resource Parse Ports to GraphQL Error", func(t *testing.T) {
		vars := []string{"foo"}
		_, err := convertPorts(vars)
		assert.EqualError(t, err, "port is not a valid integer: strconv.ParseInt: parsing \"foo\": invalid syntax")

		vars = []string{"10-9"}
		_, err = convertPorts(vars)
		assert.EqualError(t, err, "ports 10, 9 needs to be in a rising sequence")

		vars = []string{"abc-12345"}
		_, err = convertPorts(vars)
		assert.EqualError(t, err, "port is not a valid integer: strconv.ParseInt: parsing \"abc\": invalid syntax")

		vars = []string{"12345-abc"}
		_, err = convertPorts(vars)
		assert.EqualError(t, err, "port is not a valid integer: strconv.ParseInt: parsing \"abc\": invalid syntax")

		vars = []string{"1-999999"}
		_, err = convertPorts(vars)
		assert.EqualError(t, err, "port 999999 not in the range of 0-65535")

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

		err := client.createResource(resource)

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

		err := client.createResource(resource)

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

		err := client.createResource(resource)

		assert.EqualError(t, err, "failed to create resource: Post \""+client.GraphqlServerURL+"\": error_1")
	})
}

func TestClientResourceReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Client Resource Ok", func(t *testing.T) {
		// response JSON
		createResourceOkJson := `{
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
			  "policy": "ALLOW_ALL"
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
			  "policy": "RESTRICTED"
			},
			"allowIcmp": true
		  }
		}
	  }
	}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceOkJson))

		resource, err := client.readResource("resource1")
		tcpPorts, _ := resource.Protocols.TCP.buildPortsRnge()
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
		createResourceOkJson := `{
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
			  "policy": "ALLOW_ALL"
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
			  "policy": "RESTRICTED"
			},
			"allowIcmp": true
		  }
		}
	  }
	}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createResourceOkJson))

		resource, err := client.readResource("resource1")
		assert.Nil(t, resource)
		assert.EqualError(t, err, "failed to read resource with id resource1: provider does not support more than 50 groups per resource")
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

		resource, err := client.readResource("resource1")

		assert.Nil(t, resource)
		assert.EqualError(t, err, "failed to read resource with id resource1")
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

		resource, err := client.readResource(graphql.ID(""))

		assert.Nil(t, resource)
		assert.EqualError(t, err, "failed to read resource: id is empty")
	})
}

func TestClientResourceReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Read Request Error", func(t *testing.T) {
		// response JSON
		createResourceErrorJson := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, createResourceErrorJson)
				return resp, errors.New("error_1")
			})

		resource, err := client.readResource(graphql.ID("test-id"))

		assert.Nil(t, resource)
		assert.EqualError(t, err, "failed to read resource with id test-id: Post \""+client.GraphqlServerURL+"\": error_1")
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

		err := client.updateResource(resource)

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

		err := client.updateResource(resource)

		assert.EqualError(t, err, "failed to update resource with id test: cant update resource")
	})
}

func TestClientResourceUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Request Error", func(t *testing.T) {
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
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, createResourceUpdateErrorJson)
				return resp, errors.New("error_1")
			})
		resource := newTestResource()

		err := client.updateResource(resource)

		assert.EqualError(t, err, "failed to update resource with id test: Post \""+client.GraphqlServerURL+"\": error_1")
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

		err := client.deleteResource("resource1")

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

		err := client.deleteResource("resource1")

		assert.EqualError(t, err, "failed to delete resource with id resource1: cant delete resource")
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

		err := client.deleteResource("resource1")

		assert.EqualError(t, err, "failed to delete resource with id resource1: Post \""+client.GraphqlServerURL+"\": error_1")
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

		err := client.deleteResource(graphql.ID(""))

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

		edges, err := client.readResources()
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

		resources, err := client.readResources()

		assert.Nil(t, resources)
		assert.EqualError(t, err, "failed to read resource with id All: Post \""+client.GraphqlServerURL+"\": error_1")
	})
}
