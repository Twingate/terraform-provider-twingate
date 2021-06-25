package twingate

import (
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
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

	accessToken := "test1"
	refreshToken := "test2"

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, verifyTokensOkJson))
	err := client.verifyConnectorTokens(refreshToken, accessToken)

	assert.Nil(t, err)
}

// func TestClientConnectorTokensVerifyError(t *testing.T) {
// 	// response JSON
// 	verifyTokensOkJson := `{}`

// 	accessToken := "test1"
// 	refreshToken := "test2"

// 	client := newTestClient()
// 	httpmock.ActivateNonDefault(client.httpClient)
// 	defer httpmock.DeactivateAndReset()

// 	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
// 		func(req *http.Request) (*http.Response, error) {
// 			header := req.Header.Get("Authorization")
// 			assert.Contains(t, header, accessToken)
// 			resp := httpmock.NewStringResponse(501, verifyTokensOkJson)
// 			return resp, nil
// 		},
// 	)
// 	err := client.verifyConnectorTokens(refreshToken, accessToken)

// 	assert.EqualError(t, err, "failed to verify connector tokens: request  failed, status 501, body {}")
// }

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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createTokensOkJson))
	connector := &Connector{
		ID: "test",
	}
	err := client.generateConnectorTokens(connector)

	assert.EqualError(t, err, "failed to generate connector tokens with id test: error_1")
}
