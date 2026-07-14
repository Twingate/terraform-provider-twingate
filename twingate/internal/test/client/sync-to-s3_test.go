package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClientSyncToS3OidcURLReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Sync To S3 OIDC URL Read - Ok", func(t *testing.T) {
		expected := "https://tenant.twingate.com/oidc/v2"

		jsonResponse := `{
		  "data": {
		    "eventsSyncOidcProviderUrl": "https://tenant.twingate.com/oidc/v2"
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse),
		)

		url, err := c.ReadSyncToS3OidcURL(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expected, url)
	})
}

func TestClientSyncToS3OidcURLReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Sync To S3 OIDC URL Read - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		url, err := c.ReadSyncToS3OidcURL(context.Background())

		assert.Empty(t, url)
		assert.EqualError(t, err, graphqlErr(c, "failed to read sync to s3", errBadRequest))
	})
}

func TestClientSyncToS3OidcURLReadEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Sync To S3 OIDC URL Read - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "eventsSyncOidcProviderUrl": ""
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse),
		)

		url, err := c.ReadSyncToS3OidcURL(context.Background())

		assert.Empty(t, url)
		assert.NoError(t, err)
	})
}
