package twingate

import (
	b64 "encoding/base64"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func newTestResource() *Resource {
	protocols := newProcolsInput()
	protocols.TCP.Policy = graphql.String("ALLOW_ALL")
	protocols.UDP.Policy = graphql.String("ALLOW_ALL")

	groups := make([]*graphql.ID, 0)
	group := graphql.ID(b64.StdEncoding.EncodeToString([]byte("testgroup")))
	groups = append(groups, &group)

	return &Resource{
		ID:              graphql.ID("test"),
		RemoteNetworkID: graphql.ID("test"),
		Address:         graphql.String("test"),
		Name:            graphql.String("testName"),
		GroupsIds:       groups,
		Protocols:       protocols,
	}
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceOkJson))
	resource := newTestResource()

	err := client.createResource(resource)

	assert.Nil(t, err)
	assert.EqualValues(t, "test-id", resource.ID)
}

func TestClientResourceCreateError(t *testing.T) {
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceErrorJson))
	resource := newTestResource()

	err := client.createResource(resource)

	assert.EqualError(t, err, "failed to create resource with id : something went wrong")
}

func TestClientResourceReadOk(t *testing.T) {
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
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
}

func TestClientResourceReadTooManyGroups(t *testing.T) {
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceOkJson))

	resource, err := client.readResource("resource1")
	assert.Nil(t, resource)
	assert.EqualError(t, err, "failed to read resource with id resource1: provider does not support more than 50 groups per resource")
}

func TestClientResourceReadError(t *testing.T) {
	// response JSON
	createResourceErrorJson := `{
		"data": {
			"resource": null
		}
	}`

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceErrorJson))

	resource, err := client.readResource("resource1")

	assert.Nil(t, resource)
	assert.EqualError(t, err, "failed to read resource with id resource1")
}

func TestClientResourceEmptyReadError(t *testing.T) {
	// response JSON
	createResourceErrorJson := `{}`

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceErrorJson))

	resource, err := client.readResource(graphql.ID(nil))

	assert.Nil(t, resource)
	assert.EqualError(t, err, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "read", remoteNetworkResourceName, "resourceID").Error())
}

func TestClientResourceUpdateOk(t *testing.T) {
	// response JSON
	createResourceUpdateOkJson := `{
		"data": {
			"resourceUpdate": {
				"ok" : true,
				"error" : null
			}
		}
	}`

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceUpdateOkJson))
	resource := newTestResource()

	err := client.updateResource(resource)

	assert.Nil(t, err)
}

func TestClientResourceUpdateError(t *testing.T) {
	// response JSON
	createResourceUpdateErrorJson := `{
		"data": {
			"resourceUpdate": {
				"ok" : false,
				"error" : "cant update resource"
			}
		}
	}`

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceUpdateErrorJson))
	resource := newTestResource()

	err := client.updateResource(resource)

	assert.EqualError(t, err, "failed to update resource with id test: cant update resource")
}

func TestClientResourceDeleteOk(t *testing.T) {
	// response JSON
	createResourceDeleteOkJson := `{
		"data": {
			"resourceDelete": {
				"ok" : true,
				"error" : null
			}
		}
	}`

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceDeleteOkJson))

	err := client.deleteResource("resource1")

	assert.Nil(t, err)
}

func TestClientResourceDeleteError(t *testing.T) {
	// response JSON
	createResourceDeleteErrorJson := `{
		"data": {
			"resourceDelete": {
				"ok" : false,
				"error" : "cant delete resource"
			}
		}
	}`

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceDeleteErrorJson))

	err := client.deleteResource("resource1")

	assert.EqualError(t, err, "failed to delete resource with id resource1: cant delete resource")
}

func TestClientResourceEmptyDeleteError(t *testing.T) {
	// response JSON
	createResourceDeleteErrorJson := `{}`

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createResourceDeleteErrorJson))

	err := client.deleteResource(graphql.ID(nil))

	assert.EqualError(t, err, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "delete", remoteNetworkResourceName, "resourceID").Error())
}

func TestClientResourcesReadAllOk(t *testing.T) {
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, readResourcesOkJson))

	resources, err := client.readResources()
	assert.NoError(t, err)
	// Resources return dynamic and not ordered object
	// See gabs Children() method.

	r0 := &IDName{
		ID:   "resource1",
		Name: "tf-acc-resource1",
	}
	r1 := &IDName{
		ID:   "resource2",
		Name: "resource2",
	}
	r2 := &IDName{
		ID:   "resource3",
		Name: "tf-acc-resource3",
	}
	mockMap := make(map[int]*IDName)

	mockMap[0] = r0
	mockMap[1] = r1
	mockMap[2] = r2

	counter := 0
	for _, elem := range resources {
		for _, i := range mockMap {
			if elem.Node.Name == i.Name && elem.Node.ID == i.ID {
				counter++
			}
		}
	}

	if len(mockMap) != counter {
		t.Errorf("Expected map not equal to origin!")
	}
	assert.EqualValues(t, len(mockMap), counter)
}
