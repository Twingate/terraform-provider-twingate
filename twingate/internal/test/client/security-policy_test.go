package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const emptyResponse = `{}`

var requestError = errors.New("request error")

func TestClientSecurityPolicyReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Security Policy Read - Ok", func(t *testing.T) {
		expected := &model.SecurityPolicy{
			ID:   "id",
			Name: "name",
		}

		jsonResponse := `{
		  "data": {
		    "securityPolicy": {
		      "id": "id",
		      "name": "name"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse),
		)

		securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "id", "")

		assert.NoError(t, err)
		assert.Equal(t, expected, securityPolicy)
	})
}

func TestClientSecurityPolicyReadByNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Security Policy Read By Name - Ok", func(t *testing.T) {
		expected := &model.SecurityPolicy{
			ID:   "id",
			Name: "name",
		}

		jsonResponse := `{
		  "data": {
		    "securityPolicy": {
		      "id": "id",
		      "name": "name"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse),
		)

		securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "", "name")

		assert.NoError(t, err)
		assert.Equal(t, expected, securityPolicy)
	})
}

func TestClientSecurityPolicyReadWithEmptyNameAndID(t *testing.T) {
	t.Run("Test Twingate Resource : Security Policy Read With Empty Name And ID", func(t *testing.T) {
		c := newHTTPMockClient()
		securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "", "")

		assert.Nil(t, securityPolicy)
		assert.EqualError(t, err, "failed to read security policy: both name and id should not be empty")
	})
}

func TestClientSecurityPolicyReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policy - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "security-id", "")

		assert.Nil(t, securityPolicy)
		assert.EqualError(t, err, graphqlErr(c, "failed to read security policy with id security-id", errBadRequest))
	})
}

func TestClientSecurityPolicyReadEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policy - Empty Response", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "securityPolicy": null
		  }
		}`

		emptyResponses := []string{
			emptyResponse,
			response1,
		}

		c := newHTTPMockClient()

		for _, resp := range emptyResponses {
			httpmock.RegisterResponder("POST", c.GraphqlServerURL,
				httpmock.NewStringResponder(http.StatusOK, resp))

			securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "security-id", "")

			httpmock.Reset()

			assert.Nil(t, securityPolicy)
			assert.EqualError(t, err, `failed to read security policy with id security-id: query result is empty`)
		}

	})
}

func TestClientSecurityPoliciesReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Security Policies Read - Ok", func(t *testing.T) {
		expected := []*model.SecurityPolicy{
			{
				ID:   "policy-1",
				Name: "name-1",
			},
			{
				ID:   "policy-2",
				Name: "name-2",
			},
		}

		response1 := `{
		  "data": {
		    "securityPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "policy-1",
		            "name": "name-1"
		          }
		        }
		      ]
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "securityPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-2",
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "policy-2",
		            "name": "name-2"
		          }
		        }
		      ]
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
			),
		)

		securityPolicies, err := c.ReadSecurityPolicies(context.Background(), "", "")

		assert.NoError(t, err)
		assert.Equal(t, expected, securityPolicies)
	})
}

func TestClientSecurityPoliciesReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policies - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		securityPolicies, err := c.ReadSecurityPolicies(context.Background(), "", "")

		assert.Nil(t, securityPolicies)
		assert.EqualError(t, err, graphqlErr(c, "failed to read security policy", errBadRequest))
	})
}

func TestClientSecurityPoliciesReadEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policies - Empty Response", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "securityPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": false
		      },
		      "edges": []
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "securityPolicies": null
		  }
		}`

		emptyResponses := []string{
			emptyResponse,
			response1,
			response2,
		}

		c := newHTTPMockClient()

		for _, resp := range emptyResponses {
			httpmock.RegisterResponder("POST", c.GraphqlServerURL,
				httpmock.NewStringResponder(http.StatusOK, resp))

			securityPolicies, err := c.ReadSecurityPolicies(context.Background(), "", "")

			httpmock.Reset()

			assert.Nil(t, securityPolicies)
			assert.NoError(t, err)
		}

	})
}

func TestClientSecurityPoliciesReadRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policies - Request Error On Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "securityPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "policy-1",
		            "name": "name-1"
		          }
		        }
		      ]
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		securityPolicies, err := c.ReadSecurityPolicies(context.Background(), "", "")

		assert.Nil(t, securityPolicies)
		assert.EqualError(t, err, graphqlErr(c, "failed to read security policy", errBadRequest))
	})
}

func TestClientSecurityPoliciesReadEmptyResultOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policies - Empty Result on Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "securityPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "policy-1",
		            "name": "name-1"
		          }
		        }
		      ]
		    }
		  }
		}`

		response1 := `{
		  "data": {
		    "securityPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-2",
		        "hasNextPage": false
		      },
		      "edges": []
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "securityPolicies": null
		  }
		}`

		emptyResponses := []string{
			emptyResponse,
			response1,
			response2,
		}

		c := newHTTPMockClient()

		for _, resp := range emptyResponses {
			httpmock.RegisterResponder("POST", c.GraphqlServerURL,
				MultipleResponders(
					httpmock.NewStringResponder(http.StatusOK, jsonResponse),
					httpmock.NewStringResponder(http.StatusOK, resp),
				),
			)

			securityPolicies, err := c.ReadSecurityPolicies(context.Background(), "", "")

			httpmock.Reset()

			assert.Nil(t, securityPolicies)
			assert.EqualError(t, err, `failed to read security policy: query result is empty`)
		}

	})
}
