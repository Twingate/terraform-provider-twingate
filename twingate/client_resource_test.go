package twingate

// func TestParsePortsToGraphql(t *testing.T) {
// 	t.Run("Test Twingate Resource : Parse Ports to GraphQL ", func(t *testing.T) {
// 		emptyPorts, err := convertPorts(make([]string, 0))
// 		assert.NoError(t, err)
// 		assert.Equal(t, emptyPorts, "")
// 		vars := []string{"80", "81-82"}
// 		ports, err := convertPorts(vars)
// 		assert.Equal(t, ports, "{start: 80, end: 80},{start: 81, end: 82}")
// 		assert.NoError(t, err)
// 	})
// }

// func TestParseErrorPortsToGraphql(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Resource Parse Ports to GraphQL Error", func(t *testing.T) {
// 		vars := []string{"foo"}
// 		_, err := convertPorts(vars)
// 		assert.EqualError(t, err, "port is not a valid integer: strconv.ParseInt: parsing \"foo\": invalid syntax")

// 		vars = []string{"10-9"}
// 		_, err = convertPorts(vars)
// 		assert.EqualError(t, err, "ports 10, 9 needs to be in a rising sequence")

// 	})
// }

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
// 		// response JSON
// 		createResourceOkJson := `{
// 	  "data": {
// 		"resourceCreate": {
// 		  "entity": {
// 			"id": "test-id"
// 		  },
// 		  "ok": true,
// 		  "error": null
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

// 		resource := &Resource{
// 			RemoteNetworkID: "id1",
// 			Address:         "test",
// 			Name:            "testName",
// 			GroupsIds:       make([]string, 0),
// 			Protocols:       &Protocols{},
// 		}

// 		err := client.createResource(resource)

// 		assert.NoError(t, err)
// 		assert.EqualValues(t, "test-id", resource.ID)
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

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceErrorJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		resource := &Resource{
// 			RemoteNetworkID: "id1",
// 			Address:         "test",
// 			Name:            "testName",
// 			GroupsIds:       make([]string, 0),
// 			Protocols:       &Protocols{},
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
// 		// response JSON
// 		readResourcesOkJson := `{
// 	  "data": {
// 		"resources": {
// 		  "edges": [
// 			{
// 			  "node": {
// 				"id": "resource1",
// 				"name": "tf-acc-resource1"
// 			  }
// 			},
// 			{
// 			  "node": {
// 				"id": "resource2",
// 				"name": "resource2"
// 			  }
// 			},
// 			{
// 			  "node": {
// 				"id": "resource3",
// 				"name": "tf-acc-resource3"
// 			  }
// 			}
// 		  ]
// 		}
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(readResourcesOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		resources, err := client.readResources()
// 		assert.NoError(t, err)

// 		r0 := &Resources{
// 			ID:   "resource1",
// 			Name: "tf-acc-resource1",
// 		}
// 		r1 := &Resources{
// 			ID:   "resource2",
// 			Name: "resource2",
// 		}
// 		r2 := &Resources{
// 			ID:   "resource3",
// 			Name: "tf-acc-resource3",
// 		}
// 		mockMap := make(map[int]*Resources)

// 		mockMap[0] = r0
// 		mockMap[1] = r1
// 		mockMap[2] = r2

// 		counter := 0
// 		for _, elem := range resources {
// 			for _, i := range mockMap {
// 				if elem.Name == i.Name && elem.ID == i.ID {
// 					counter++
// 				}
// 			}
// 		}

// 		if len(mockMap) != counter {
// 			t.Errorf("Expected map not equal to origin!")
// 		}
// 		assert.EqualValues(t, len(mockMap), counter)
// 	})
// }
