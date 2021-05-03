package twingate

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
)

func TestParsePortsToGraphql(t *testing.T) {
	emptyPorts := convertPorts(make([]string, 0))
	assert.Equal(t, emptyPorts, "")
	vars := []string{"80", "81-82"}
	ports := convertPorts(vars)
	assert.Equal(t, ports, "{start: 80, end: 80},{start: 81, end: 82}")
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

	r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	resource := &Resource{
		RemoteNetworkId: "id1",
		Address:         "test",
		Name:            "testName",
		Groups:          make([]string, 0),
		Protocols:       &Protocols{},
	}

	err := client.createResource(resource)

	assert.Nil(t, err)
	assert.EqualValues(t, "test-id", resource.Id)
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

	r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceErrorJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	resource := &Resource{
		RemoteNetworkId: "id1",
		Address:         "test",
		Name:            "testName",
		Groups:          make([]string, 0),
		Protocols:       &Protocols{},
	}

	err := client.createResource(resource)

	assert.EqualError(t, err, "api request error : can't create resource name testName, error: something went wrong")
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

	r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	resource, err := client.readResource("resource1")

	assert.Nil(t, err)
	assert.EqualValues(t, "resource1", resource.Id)
	assert.Contains(t, resource.Groups, "group1")
	assert.Contains(t, resource.Protocols.TCPPorts, "8080-8090")
	assert.EqualValues(t, resource.Address, "test.com")
	assert.EqualValues(t, resource.RemoteNetworkId, "network1")
	assert.Len(t, resource.Protocols.UDPPorts, 0)
	assert.EqualValues(t, resource.Name, "test resource")
}

func TestClientResourceReadError(t *testing.T) {

	// response JSON
	createResourceErrorJson := `{
		"data": {
			"resource": null
		}
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceErrorJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	resource, err := client.readResource("resource1")

	assert.Nil(t, resource)
	assert.EqualError(t, err, "api request error : can't read resource: resource1")
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

	r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceUpdateOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	resource := &Resource{
		RemoteNetworkId: "network1",
		Address:         "test.com",
		Name:            "test resource",
		Groups:          make([]string, 0),
		Protocols:       &Protocols{},
	}

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

	r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceUpdateErrorJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	resource := &Resource{
		RemoteNetworkId: "network1",
		Address:         "test.com",
		Name:            "test resource",
		Groups:          make([]string, 0),
		Protocols:       &Protocols{},
	}

	err := client.updateResource(resource)

	assert.EqualError(t, err, "api request error : can't update resource: cant update resource")
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

	r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceDeleteOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

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

	r := ioutil.NopCloser(bytes.NewReader([]byte(createResourceDeleteErrorJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	err := client.deleteResource("resource1")

	assert.EqualError(t, err, "api request error : unable to delete resource Id resource1, error: cant delete resource")
}
