package client

import (
	"context"
	"fmt"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClientDLPPolicyReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policy Ok", func(t *testing.T) {
		expected := &model.DLPPolicy{
			ID:   "policy-id",
			Name: "policy-name",
		}

		jsonResponse := `{
		  "data": {
		    "dlpPolicy": {
		      "id": "policy-id",
		      "name": "policy-name"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		policy, err := c.ReadDLPPolicy(context.Background(), &model.DLPPolicy{ID: "policy-id"})

		assert.NoError(t, err)
		assert.Equal(t, expected, policy)
	})
}

func TestClientDLPPolicyReadOkQueryByName(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policy Ok Query By Name", func(t *testing.T) {
		expected := &model.DLPPolicy{
			ID:   "policy-id",
			Name: "policy-name",
		}

		jsonResponse := `{
		  "data": {
		    "dlpPolicy": {
		      "id": "policy-id",
		      "name": "policy-name"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		policy, err := c.ReadDLPPolicy(context.Background(), &model.DLPPolicy{Name: "policy-name"})

		assert.NoError(t, err)
		assert.Equal(t, expected, policy)
	})
}

func TestClientDLPPolicyReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policy Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dlpPolicy": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const policyID = "policy-id"
		policy, err := c.ReadDLPPolicy(context.Background(), &model.DLPPolicy{ID: policyID})

		assert.Nil(t, policy)
		assert.EqualError(t, err, fmt.Sprintf("failed to read dlp policy with id %s: query result is empty", policyID))
	})
}

func TestClientDLPPolicyReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policy Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		const policyID = "policy-id"
		policy, err := c.ReadDLPPolicy(context.Background(), &model.DLPPolicy{ID: policyID})

		assert.Nil(t, policy)
		assert.EqualError(t, err, graphqlErr(c, "failed to read dlp policy with id "+policyID, errBadRequest))
	})
}

func TestClientReadEmptyDLPPolicyErrorWithNullPolicy(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty DLP Policy Error with null policy", func(t *testing.T) {

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		policy, err := c.ReadDLPPolicy(context.Background(), nil)

		assert.EqualError(t, err, "failed to read dlp policy: both name and id should not be empty")
		assert.Nil(t, policy)
	})
}

func TestClientReadEmptyDLPPolicyError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty DLP Policy Error", func(t *testing.T) {

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		policy, err := c.ReadDLPPolicy(context.Background(), &model.DLPPolicy{})

		assert.EqualError(t, err, "failed to read dlp policy: both name and id should not be empty")
		assert.Nil(t, policy)
	})
}

func TestClientDLPPoliciesReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policies Ok", func(t *testing.T) {
		expected := []*model.DLPPolicy{
			{
				ID:   "id1",
				Name: "policy1",
			},
			{
				ID:   "id2",
				Name: "policy2",
			},
			{
				ID:   "id3",
				Name: "policy3",
			},
		}

		jsonResponse := `{
		  "data": {
		    "dlpPolicies": {
		      "edges": [
		        {
		          "node": {
		            "id": "id1",
		            "name": "policy1"
		          }
		        },
		        {
		          "node": {
		            "id": "id2",
		            "name": "policy2"
		          }
		        },
		        {
		          "node": {
		            "id": "id3",
		            "name": "policy3"
		          }
		        }
		      ]
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		policies, err := c.ReadDLPPolicies(context.Background(), "policy", "_prefix")

		assert.NoError(t, err)
		assert.Equal(t, expected, policies)
	})
}

func TestClientDLPPoliciesReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policies Error", func(t *testing.T) {
		emptyResponse := `{
		  "data": {
		    "dlpPolicies": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, emptyResponse))

		policies, err := c.ReadDLPPolicies(context.Background(), "", "")

		assert.Nil(t, policies)
		assert.EqualError(t, err, "failed to read dlp policy with id All: query result is empty")
	})
}

func TestClientDLPPoliciesReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policies Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		policies, err := c.ReadDLPPolicies(context.Background(), "", "")

		assert.Nil(t, policies)
		assert.EqualError(t, err, graphqlErr(c, "failed to read dlp policy with id All", errBadRequest))
	})
}

func TestClientDLPPoliciesReadRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policies - Request Error on Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dlpPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-001",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id1",
		            "name": "policy1"
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
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		policies, err := c.ReadDLPPolicies(context.Background(), "policy", "_regexp")

		assert.Nil(t, policies)
		assert.EqualError(t, err, graphqlErr(c, "failed to read dlp policy with id All", errBadRequest))
	})
}

func TestClientDLPPoliciesReadEmptyResultOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policies - Empty Result on Fetching", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "dlpPolicies": {
			"pageInfo": {
		        "endCursor": "cursor-001",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id1",
		            "name": "policy1"
		          }
		        }
		      ]
		    }
		  }
		}`

		response2 := `{}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, response1),
				httpmock.NewStringResponder(200, response2),
			),
		)

		policies, err := c.ReadDLPPolicies(context.Background(), "policy1", "_suffix")

		assert.Nil(t, policies)
		assert.EqualError(t, err, `failed to read dlp policy with id All: query result is empty`)
	})
}

func TestClientDLPPoliciesReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read DLP Policies All - Ok", func(t *testing.T) {
		expected := []*model.DLPPolicy{
			{ID: "id-1", Name: "policy-1"},
			{ID: "id-2", Name: "policy-2"},
			{ID: "id-3", Name: "policy-3"},
		}

		jsonResponse := `{
		  "data": {
		    "dlpPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-001",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "policy-1"
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "policy-2"
		          }
		        }
		      ]
		    }
		  }
		}`

		nextPage := `{
		  "data": {
		    "dlpPolicies": {
		      "pageInfo": {
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-3",
		            "name": "policy-3"
		          }
		        }
		      ]
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.ResponderFromMultipleResponses(
				[]*http.Response{
					httpmock.NewStringResponse(200, jsonResponse),
					httpmock.NewStringResponse(200, nextPage),
				}),
		)

		policies, err := c.ReadDLPPolicies(context.Background(), "policy", "_contains")

		assert.NoError(t, err)
		assert.Equal(t, expected, policies)
	})
}
