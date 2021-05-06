package twingate

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/stretchr/testify/assert"
)

func TestClientConnectorCreateTokensOK(t *testing.T) {
	// response JSON
	createTokensOkJson := `{
		"data": {
			"connectorGenerateTokens": {
				"connectorTokens": {
					"accessToken": "token1",
					"refreshToken": "token2"
				},
				"ok": true,
				"error": null
			}
		}
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(createTokensOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	connector := &Connector{
		Id: "test",
	}
	err := client.generateConnectorTokens(connector)

	assert.Nil(t, err)
	assert.EqualValues(t, "token1", connector.ConnectorTokens.AccessToken)
	assert.EqualValues(t, "token2", connector.ConnectorTokens.RefreshToken)
}

func TestClientConnectorTokensVerifyOK(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(verifyTokensOkJson)))
	client := createTestClient()

	accessToken := "test1"
	refreshToken := "test2"

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		header := req.Header.Get("Authorization")
		assert.Contains(t, header, accessToken)
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	err := client.verifyConnectorTokens(refreshToken, accessToken)

	assert.Nil(t, err)
}

func TestClientConnectorTokensVerifyError(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(verifyTokensOkJson)))
	client := createTestClient()

	accessToken := "test1"
	refreshToken := "test2"

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		header := req.Header.Get("Authorization")
		assert.Contains(t, header, accessToken)
		return &http.Response{
			StatusCode: 501,
			Body:       r,
		}, nil
	}
	err := client.verifyConnectorTokens(refreshToken, accessToken)

	assert.EqualError(t, err, "failed to verify connector tokens: request  failed, status 501, body {}")
}

func TestClientConnectorCreateTokensError(t *testing.T) {
	// response JSON
	createTokensOkJson := `{
	  "data": {
		"connectorGenerateTokens": {
		  "ok": false,
		  "error": "error_1"
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(createTokensOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	connector := &Connector{
		Id: "test",
	}
	err := client.generateConnectorTokens(connector)

	assert.EqualError(t, err, "failed to generate connector tokens with id test: error_1")
}
