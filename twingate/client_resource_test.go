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
	emptyPorts := convertPortsToGraphql(make([]string, 0))
	assert.Equal(t, emptyPorts, "")
	vars := []string{"80", "81-82"}
	ports := convertPortsToGraphql(vars)
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
