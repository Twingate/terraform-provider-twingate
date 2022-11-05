package client

import (
	"context"
	"errors"
	"fmt"
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

	tokens, err := client.GenerateConnectorTokens(context.Background(), "connector-id")

	assert.Nil(t, err)
	assert.EqualValues(t, "token1", tokens.AccessToken)
	assert.EqualValues(t, "token2", tokens.RefreshToken)
}

func TestClientConnectorTokensVerifyOK(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()

	accessToken := "test1"
	refreshToken := "test2"

	httpmock.RegisterResponder("POST", client.APIServerURL+"/connector/validate_tokens",
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return httpmock.NewStringResponse(200, verifyTokensOkJson), nil
		})

	err := client.VerifyConnectorTokens(context.Background(), refreshToken, accessToken)

	assert.Nil(t, err)
}

func TestClientConnectorTokensVerify401Error(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()

	accessToken := "test1"
	refreshToken := "test2"

	apiURL := client.APIServerURL + "/connector/validate_tokens"
	httpmock.RegisterResponder("POST", apiURL,
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return httpmock.NewStringResponse(401, verifyTokensOkJson), nil
		})

	err := client.VerifyConnectorTokens(context.Background(), refreshToken, accessToken)

	assert.EqualError(t, err, fmt.Sprintf("failed to verify connector tokens: request %s failed, status 401, body {}", apiURL))
}

func TestClientConnectorTokensVerifyRequestError(t *testing.T) {
	client := newHTTPMockClient()

	accessToken := "test1"
	refreshToken := "test2"

	defer httpmock.DeactivateAndReset()
	apiURL := client.APIServerURL + "/connector/validate_tokens"
	httpmock.RegisterResponder("POST", apiURL,
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return nil, errors.New("error")
		})

	err := client.VerifyConnectorTokens(context.Background(), refreshToken, accessToken)
	assert.EqualError(t, err, fmt.Sprintf(`failed to verify connector tokens: can't execute http request: Post "%s": error`, apiURL))
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

	const connectorID = "test-id"
	_, err := client.GenerateConnectorTokens(context.Background(), connectorID)

	assert.EqualError(t, err, fmt.Sprintf(`failed to generate connector tokens with id %v: error_1`, connectorID))
}

func TestClientConnectorTokensCreateRequestError(t *testing.T) {
	client := newHTTPMockClient()

	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewErrorResponder(errors.New("error_1")))

	_, err := client.GenerateConnectorTokens(context.Background(), "connector-id")

	assert.EqualError(t, err, fmt.Sprintf(`failed to generate connector tokens: Post "%s": error_1`, client.GraphqlServerURL))
}
