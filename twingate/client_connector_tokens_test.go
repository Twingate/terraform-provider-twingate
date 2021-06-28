package twingate

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
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

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createTokensOkJson))
	connector := &Connector{
		ID: "test",
	}
	err := client.generateConnectorTokens(connector)

	assert.Nil(t, err)
	assert.EqualValues(t, "token1", connector.ConnectorTokens.AccessToken)
	assert.EqualValues(t, "token2", connector.ConnectorTokens.RefreshToken)
}

func TestClientConnectorTokensVerifyOK(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()

	accessToken := "test1"
	refreshToken := "test2"

	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return httpmock.NewStringResponse(200, verifyTokensOkJson), nil
		})

	err := client.verifyConnectorTokens(refreshToken, accessToken)

	assert.Nil(t, err)
}

func TestClientConnectorTokensVerifyError(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()

	accessToken := "test1"
	refreshToken := "test2"

	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return httpmock.NewStringResponse(501, verifyTokensOkJson), nil
		})

	err := client.verifyConnectorTokens(refreshToken, accessToken)

	assert.EqualError(t, err, "failed to verify connector tokens with id : request  failed, status 501, body {}")
}

func TestClientConnectorTokensRequestError(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	client := newHTTPMockClient()

	accessToken := "test1"
	refreshToken := "test2"
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return httpmock.NewStringResponse(501, verifyTokensOkJson), errors.New("error")
		})

	err := client.verifyConnectorTokens(refreshToken, accessToken)

	assert.EqualError(t, err, "failed to verify connector tokens with id : can't execute http request: error")
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

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createTokensOkJson))
	connector := &Connector{
		ID: "test",
	}
	err := client.generateConnectorTokens(connector)

	assert.EqualError(t, err, "failed to generate connector tokens with id test: error_1")
}

func TestClientConnectorCreateTokensRequestError(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`
	connector := &Connector{ID: "test"}

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, verifyTokensOkJson)
			return resp, errors.New("error")
		})

	err := client.generateConnectorTokens(connector)

	assert.EqualError(t, err, "failed to generate connector tokens with id : Post \""+client.GraphqlServerURL+"\": error")
}

func TestClientConnectorEmptyCreateTokensError(t *testing.T) {
	// response JSON
	createTokensOkJson := `{}`
	connector := &Connector{ID: nil}

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createTokensOkJson))

	err := client.generateConnectorTokens(connector)

	assert.EqualError(t, err, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "generate", remoteNetworkResourceName, "connectorTokens").Error())
}
