package twingate

import (
	b64 "encoding/base64"
	"terraform-provider-twingate/mock_twingate"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

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

	})
}

// func TestConvertProtocolsErrors(t *testing.T) {
// 	t.Run("Test Twingate Resource : Convert Protocols Errors", func(t *testing.T) {
// 		var protocols *Protocols
// 		p, err := convertProtocols(protocols)
// 		assert.EqualValues(t, "", p)
// 		assert.NoError(t, err)
// 	})
// }

// func TestClientResourceCreateOk(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Create Ok", func(t *testing.T) {
// 		protocols := newProcolsInput()
// 		protocols.TCP.Policy = graphql.String("ALLOW_ALL")
// 		protocols.UDP.Policy = graphql.String("ALLOW_ALL")

// 		groups := make([]*graphql.ID, 0)
// 		group := graphql.ID("testgroup")
// 		groups = append(groups, &group)

// 		resource := &Resource{
// 			RemoteNetworkID: graphql.ID("testmeplease"),
// 			Address:         graphql.String("test"),
// 			Name:            graphql.String("testName"),
// 			GroupsIds:       groups,
// 			Protocols:       protocols,
// 		}

// 		client, _ := sharedClient("terraformtests")
// 		err := client.createResource(resource)

// 		assert.NoError(t, err)
// 		assert.EqualValues(t, graphql.ID("test-id"), resource.ID)
// 	})
// }

// func TestClientResourceCreateError(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Create Error", func(t *testing.T) {
// 		// response JSON
// 		createResourceErrorJson := `{
// 	  "data": {
// 		"resourceCreate": {
// 		  "entity": {
// 			"id": "test-id"
// 		  },
// 		  "ok": false,
// 		  "error": "something went wrong"
// 		}
// 	  }
// 	}`

// 		client, _ := sharedClient("terraformtests")

// 		resource := &Resource{
// 			RemoteNetworkID: "id1",
// 			Address:         "test",
// 			Name:            "testName",
// 			GroupsIds:       make([]*graphql.ID, 0),
// 			Protocols:       &ProtocolsInput{},
// 		}

// 		err := client.createResource(resource)

// 		assert.EqualError(t, err, "failed to create resource: something went wrong")
// 	})
// }

// func TestClientResourceReadOk(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Read Ok", func(t *testing.T) {
// 		// response JSON
// 		createResourceOkJson := `{
// 	  "data": {
// 		"resource": {
// 		  "id": "resource1",
// 		  "name": "test resource",
// 		  "address": {
// 			"type": "DNS",
// 			"value": "test.com"
// 		  },
// 		  "remoteNetwork": {
// 			"id": "network1"
// 		  },
// 		  "groups": {
// 			"pageInfo": {
// 			  "hasNextPage": false
// 			},
// 			"edges": [
// 			  {
// 				"node": {
// 				  "id": "group1"
// 				}
// 			  },
// 			  {
// 				"node": {
// 				  "id": "group2"
// 				}
// 			  }
// 			]
// 		  },
// 		  "protocols": {
// 			"udp": {
// 			  "ports": [],
// 			  "policy": "ALLOW_ALL"
// 			},
// 			"tcp": {
// 			  "ports": [
// 				{
// 				  "end": 80,
// 				  "start": 80
// 				},
// 				{
// 				  "end": 8090,
// 				  "start": 8080
// 				}
// 			  ],
// 			  "policy": "RESTRICTED"
// 			},
// 			"allowIcmp": true
// 		  }
// 		}
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		resource, err := client.readResource("resource1")

// 		assert.NoError(t, err)
// 		assert.EqualValues(t, "resource1", resource.ID)
// 		assert.Contains(t, resource.GroupsIds, "group1")
// 		assert.Contains(t, resource.Protocols.TCPPorts, "8080-8090")
// 		assert.EqualValues(t, resource.Address, "test.com")
// 		assert.EqualValues(t, resource.RemoteNetworkID, "network1")
// 		assert.Len(t, resource.Protocols.UDPPorts, 0)
// 		assert.EqualValues(t, resource.Name, "test resource")
// 	})
// }

// func TestClientResourceReadTooManyGroups(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Read Too Many Groups", func(t *testing.T) {
// 		// response JSON
// 		createResourceOkJson := `{
// 	  "data": {
// 		"resource": {
// 		  "id": "resource1",
// 		  "name": "test resource",
// 		  "address": {
// 			"type": "DNS",
// 			"value": "test.com"
// 		  },
// 		  "remoteNetwork": {
// 			"id": "network1"
// 		  },
// 		  "groups": {
// 			"pageInfo": {
// 			  "hasNextPage": true
// 			},
// 			"edges": [
// 			  {
// 				"node": {
// 				  "id": "group1"
// 				}
// 			  },
// 			  {
// 				"node": {
// 				  "id": "group2"
// 				}
// 			  }
// 			]
// 		  },
// 		  "protocols": {
// 			"udp": {
// 			  "ports": [],
// 			  "policy": "ALLOW_ALL"
// 			},
// 			"tcp": {
// 			  "ports": [
// 				{
// 				  "end": 80,
// 				  "start": 80
// 				},
// 				{
// 				  "end": 8090,
// 				  "start": 8080
// 				}
// 			  ],
// 			  "policy": "RESTRICTED"
// 			},
// 			"allowIcmp": true
// 		  }
// 		}
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		resource, err := client.readResource("resource1")
// 		assert.Nil(t, resource)
// 		assert.EqualError(t, err, "failed to read resource with id resource1: provider does not support more than 50 groups per resource")
// 	})
// }

// func TestClientResourceReadError(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Read Error", func(t *testing.T) {
// 		// response JSON
// 		createResourceErrorJson := `{
// 		"data": {
// 			"resource": null
// 		}
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceErrorJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		resource, err := client.readResource("resource1")

// 		assert.Nil(t, resource)
// 		assert.EqualError(t, err, "failed to read resource with id resource1")
// 	})
// }

func TestClientResourceUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Ok", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		gqlMock := mock_twingate.NewMockGql(mockCtrl)
		protocols := newProcolsInput()
		protocols.TCP.Policy = graphql.String("ALLOW_ALL")
		protocols.UDP.Policy = graphql.String("ALLOW_ALL")

		groups := make([]*graphql.ID, 0)
		group := graphql.ID(b64.StdEncoding.EncodeToString([]byte("testgroup")))
		groups = append(groups, &group)

		resource := &Resource{
			ID:              graphql.ID("test"),
			RemoteNetworkID: graphql.ID("test"),
			Address:         graphql.String("test"),
			Name:            graphql.String("testName"),
			GroupsIds:       groups,
			Protocols:       protocols,
		}
		f := func() Gql {
			r := updateResourceQuery{}

			variables := map[string]interface{}{
				"id":              resource.ID,
				"name":            resource.Name,
				"address":         resource.Address,
				"remoteNetworkId": resource.RemoteNetworkID,
				"groupIds":        resource.GroupsIds,
				"protocols":       resource.Protocols,
			}

			v := updateResourceQuery{
				ResourceUpdate: &OkError{
					Ok: graphql.Boolean(true),
				},
			}

			gqlMock.EXPECT().Mutate(gomock.Any(), &r, variables).SetArg(1, v).Return(nil).Times(1)
			return gqlMock
		}

		c := Client{GraphqlClient: f()}

		err := c.updateResource(resource)
		assert.NoError(t, err)
	})
}

// func TestClientResourceUpdateOk(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Update Ok", func(t *testing.T) {
// 		// response JSON
// 		createResourceUpdateOkJson := `{
// 		"data": {
// 			"resourceUpdate": {
// 				"ok" : true,
// 				"error" : null
// 			}
// 		}
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceUpdateOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		resource := &Resource{
// 			RemoteNetworkID: "network1",
// 			Address:         "test.com",
// 			Name:            "test resource",
// 			GroupsIds:       make([]string, 0),
// 			Protocols:       &Protocols{},
// 		}

// 		err := client.updateResource(resource)

// 		assert.NoError(t, err)
// 	})
// }

// func TestClientResourceUpdateError(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Update Error", func(t *testing.T) {
// 		// response JSON
// 		createResourceUpdateErrorJson := `{
// 		"data": {
// 			"resourceUpdate": {
// 				"ok" : false,
// 				"error" : "cant update resource"
// 			}
// 		}
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceUpdateErrorJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		resource := &Resource{
// 			RemoteNetworkID: "network1",
// 			Address:         "test.com",
// 			Name:            "test resource",
// 			GroupsIds:       make([]string, 0),
// 			Protocols:       &Protocols{},
// 		}

// 		err := client.updateResource(resource)

// 		assert.EqualError(t, err, "failed to update resource: cant update resource")
// 	})
// }

func TestClientResourceUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Update Error", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		gqlMock := mock_twingate.NewMockGql(mockCtrl)

		protocols := newProcolsInput()
		protocols.TCP.Policy = graphql.String("ALLOW_ALL")
		protocols.UDP.Policy = graphql.String("ALLOW_ALL")

		groups := make([]*graphql.ID, 0)
		group := graphql.ID(b64.StdEncoding.EncodeToString([]byte("testgroup")))
		groups = append(groups, &group)

		resource := &Resource{
			ID:              graphql.ID("test"),
			RemoteNetworkID: graphql.ID("test"),
			Address:         graphql.String("test"),
			Name:            graphql.String("testName"),
			GroupsIds:       groups,
			Protocols:       protocols,
		}

		f := func() Gql {
			r := updateResourceQuery{}

			variables := map[string]interface{}{
				"id":              resource.ID,
				"name":            resource.Name,
				"address":         resource.Address,
				"remoteNetworkId": resource.RemoteNetworkID,
				"groupIds":        resource.GroupsIds,
				"protocols":       resource.Protocols,
			}

			v := updateResourceQuery{
				ResourceUpdate: &OkError{
					Ok:    graphql.Boolean(false),
					Error: graphql.String("cant update resource"),
				},
			}

			gqlMock.EXPECT().Mutate(gomock.Any(), &r, variables).SetArg(1, v).Return(nil).Times(1)
			return gqlMock
		}

		c := Client{GraphqlClient: f()}

		err := c.updateResource(resource)
		assert.EqualError(t, err, "failed to update resource with id "+resource.ID.(string)+": cant update resource")
	})
}

// func TestClientResourceDeleteOk(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Delete Ok", func(t *testing.T) {
// 		// response JSON
// 		createResourceDeleteOkJson := `{
// 		"data": {
// 			"resourceDelete": {
// 				"ok" : true,
// 				"error" : null
// 			}
// 		}
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceDeleteOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		err := client.deleteResource("resource1")

// 		assert.NoError(t, err)
// 	})
// }

func TestClientResourceDeleteOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Resource Delete Ok", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		gqlMock := mock_twingate.NewMockGql(mockCtrl)

		f := func() Gql {
			r := deleteResourceQuery{}

			variables := map[string]interface{}{
				"id": graphql.ID("test"),
			}

			v := deleteResourceQuery{
				ResourceDelete: &OkError{
					Ok: graphql.Boolean(true),
				},
			}

			gqlMock.EXPECT().Mutate(gomock.Any(), &r, variables).SetArg(1, v).Return(nil).Times(1)
			return gqlMock
		}

		c := Client{GraphqlClient: f()}

		err := c.deleteResource(graphql.ID("test"))
		assert.NoError(t, err)
	})
}

// func TestClientResourceDeleteError(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Delete Error", func(t *testing.T) {
// 		// response JSON
// 		createResourceDeleteErrorJson := `{
// 		"data": {
// 			"resourceDelete": {
// 				"ok" : false,
// 				"error" : "cant delete resource"
// 			}
// 		}
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceDeleteErrorJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		err := client.deleteResource("resource1")

// 		assert.EqualError(t, err, "failed to delete resource with id resource1: cant delete resource")
// 	})
// }

// func TestClientResourcesReadAllOk(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Read All Ok", func(t *testing.T) {
// 		client, _ := sharedClient("terraformtests")
// 		resources := readResourcesQuery{}
// 		variables := map[string]interface{}{}

// 		edges := []*Edges{}

// 		r0 := &Edges{&IDName{
// 			ID:   "resource1",
// 			Name: "tf-acc-resource1",
// 		}}
// 		r1 := &Edges{&IDName{
// 			ID:   "resource2",
// 			Name: "resource2",
// 		}}
// 		r2 := &Edges{&IDName{
// 			ID:   "resource3",
// 			Name: "tf-acc-resource3",
// 		}}

// 		edges = append(edges, r0)
// 		edges = append(edges, r1)
// 		edges = append(edges, r2)
// 		resources.Resources.Edges = edges

// 		err := client.GraphqlClient.Query(context.Background(), &resources, variables)
// 		assert.NoError(t, err)

// 		mockMap := make(map[graphql.ID]graphql.String)

// 		mockMap[r0.Node.ID] = r0.Node.Name
// 		mockMap[r1.Node.ID] = r1.Node.Name
// 		mockMap[r2.Node.ID] = r2.Node.Name

// 		for _, elem := range resources.Resources.Edges {
// 			name := mockMap[elem.Node.ID]
// 			assert.Equal(t, name, elem.Node.Name)
// 		}
// 	})
// }
