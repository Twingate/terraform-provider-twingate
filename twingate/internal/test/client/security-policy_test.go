package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const emptyResponse = `{}`

var requestError = errors.New("request error")

func TestClientSecurityPolicyReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Security Policy Read - Ok", func(t *testing.T) {
		expected := &model.SecurityPolicy{
			ID:     "id",
			Name:   "name",
			Type:   "DEFAULT_RESOURCE",
			Groups: []string{"g-1", "g-2"},
		}

		jsonResponse := `{
		  "data": {
		    "securityPolicy": {
		      "id": "id",
		      "name": "name",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": true
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "g-1",
		              "name": "group1"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		nextResponse := `{
		  "data": {
		    "securityPolicy": {
		      "id": "id",
		      "name": "name",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "g-2",
		              "name": "group2"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewStringResponder(http.StatusOK, nextResponse),
			),
		)

		securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "id", "")

		assert.NoError(t, err)
		assert.Equal(t, expected, securityPolicy)
	})
}

func TestClientSecurityPolicyReadByNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Security Policy Read By Name - Ok", func(t *testing.T) {
		expected := &model.SecurityPolicy{
			ID:     "id",
			Name:   "name",
			Type:   "DEFAULT_RESOURCE",
			Groups: []string{"g-1", "g-2"},
		}

		jsonResponse := `{
		  "data": {
		    "securityPolicy": {
		      "id": "id",
		      "name": "name",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": true
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "g-1",
		              "name": "group1"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		nextResponse := `{
		  "data": {
		    "securityPolicy": {
		      "id": "id",
		      "name": "name",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "g-2",
		              "name": "group2"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewStringResponder(http.StatusOK, nextResponse),
			),
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
			httpmock.NewErrorResponder(requestError))

		securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "security-id", "")

		assert.Nil(t, securityPolicy)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read security policy with id security-id: Post "%s": request error`, c.GraphqlServerURL))
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

func TestClientSecurityPolicyReadRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policy - Request Error On Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "securityPolicy": {
		      "id": "security-id",
		      "name": "name",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": true
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "g-1",
		              "name": "group1"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewErrorResponder(requestError),
			),
		)

		securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "security-id", "")

		assert.Nil(t, securityPolicy)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read security policy with id security-id: failed to read group with id All: Post "%s": request error`, c.GraphqlServerURL))
	})
}

func TestClientSecurityPolicyReadEmptyResultOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policy - Empty Result on Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "securityPolicy": {
		      "id": "security-id",
		      "name": "name",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": true
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "g-1",
		              "name": "group1"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		response1 := `{
		  "data": {
		    "securityPolicy": {
		      "id": "security-id",
		      "name": "name",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "cursor-001",
		          "hasNextPage": false
		        },
		        "edges": []
		      }
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "securityPolicy": {
		      "id": "security-id",
		      "name": "name",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": null
		    }
		  }
		}`

		response3 := `{
		  "data": {
		    "securityPolicy": null
		  }
		}`

		emptyResponses := []string{
			emptyResponse,
			response1,
			response2,
			response3,
		}

		c := newHTTPMockClient()

		for _, resp := range emptyResponses {
			httpmock.RegisterResponder("POST", c.GraphqlServerURL,
				MultipleResponders(
					httpmock.NewStringResponder(http.StatusOK, jsonResponse),
					httpmock.NewStringResponder(http.StatusOK, resp),
				),
			)

			securityPolicy, err := c.ReadSecurityPolicy(context.Background(), "security-id", "")

			httpmock.Reset()

			assert.Nil(t, securityPolicy)
			assert.EqualError(t, err, `failed to read security policy with id security-id: failed to read group with id All: query result is empty`)
		}
	})
}

func TestClientSecurityPoliciesReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Security Policies Read - Ok", func(t *testing.T) {
		expected := []*model.SecurityPolicy{
			{
				ID:     "policy-1",
				Name:   "name-1",
				Type:   "DEFAULT_RESOURCE",
				Groups: []string{"g-1", "g-2"},
			},
			{
				ID:     "policy-2",
				Name:   "name-2",
				Type:   "DEFAULT_RESOURCE",
				Groups: []string{"g-3", "g-4"},
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
		            "name": "name-1",
		            "policyType": "DEFAULT_RESOURCE",
		            "groups": {
		              "pageInfo": {
		                "endCursor": "groups-cursor-1",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "g-1",
		                    "name": "group1"
		                  }
		                }
		              ]
		            }
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
		            "name": "name-2",
		            "policyType": "DEFAULT_RESOURCE",
		            "groups": {
		              "pageInfo": {
		                "endCursor": "groups-cursor-2",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "g-3",
		                    "name": "group3"
		                  }
		                }
		              ]
		            }
		          }
		        }
		      ]
		    }
		  }
		}`

		response3 := `{
		  "data": {
		    "securityPolicy": {
		      "id": "policy-1",
		      "name": "name-1",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "groups-cursor-1-end",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "g-2",
		              "name": "group2"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		response4 := `{
		  "data": {
		    "securityPolicy": {
		      "id": "policy-2",
		      "name": "name-2",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "groups-cursor-2-end",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "g-4",
		              "name": "group4"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
				httpmock.NewStringResponder(http.StatusOK, response3),
				httpmock.NewStringResponder(http.StatusOK, response4),
			),
		)

		securityPolicies, err := c.ReadSecurityPolicies(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expected, securityPolicies)
	})
}

func TestClientSecurityPoliciesReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policies - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(requestError))

		securityPolicies, err := c.ReadSecurityPolicies(context.Background())

		assert.Nil(t, securityPolicies)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read security policy: Post "%s": request error`, c.GraphqlServerURL))
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

			securityPolicies, err := c.ReadSecurityPolicies(context.Background())

			httpmock.Reset()

			assert.Nil(t, securityPolicies)
			assert.EqualError(t, err, `failed to read security policy: query result is empty`)
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
		            "name": "name-1",
		            "policyType": "DEFAULT_RESOURCE",
		            "groups": null
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
				httpmock.NewErrorResponder(requestError),
			),
		)

		securityPolicies, err := c.ReadSecurityPolicies(context.Background())

		assert.Nil(t, securityPolicies)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read security policy: Post "%s": request error`, c.GraphqlServerURL))
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
		            "name": "name-1",
		            "policyType": "DEFAULT_RESOURCE",
		            "groups": null
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

			securityPolicies, err := c.ReadSecurityPolicies(context.Background())

			httpmock.Reset()

			assert.Nil(t, securityPolicies)
			assert.EqualError(t, err, `failed to read security policy: query result is empty`)
		}

	})
}

func TestClientSecurityPoliciesReadRequestErrorOnFetchingGroups(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policies - Request Error On Fetching Groups", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "securityPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "policy-1",
		            "name": "name-1",
		            "policyType": "DEFAULT_RESOURCE",
		            "groups": {
		              "pageInfo": {
		                "endCursor": "groups-cursor-1",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "g-1",
		                    "name": "group1"
		                  }
		                }
		              ]
		            }
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
				httpmock.NewErrorResponder(requestError),
			),
		)

		securityPolicies, err := c.ReadSecurityPolicies(context.Background())

		assert.Nil(t, securityPolicies)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read security policy with id policy-1: failed to read group with id All: Post "%s": request error`, c.GraphqlServerURL))
	})
}

func TestClientSecurityPoliciesReadEmptyResultOnFetchingGroups(t *testing.T) {
	t.Run("Test Twingate Resource : Read Security Policies - Empty Result on Fetching Groups", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "securityPolicies": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "policy-1",
		            "name": "name-1",
		            "policyType": "DEFAULT_RESOURCE",
		            "groups": {
		              "pageInfo": {
		                "endCursor": "groups-cursor-1",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "g-1",
		                    "name": "group1"
		                  }
		                }
		              ]
		            }
		          }
		        }
		      ]
		    }
		  }
		}`

		response1 := `{
		  "data": {
		    "securityPolicy": {
		      "id": "policy-1",
		      "name": "name-1",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": {
		        "pageInfo": {
		          "endCursor": "groups-cursor-1-end",
		          "hasNextPage": false
		        },
		        "edges": []
		      }
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "securityPolicy": {
		      "id": "policy-1",
		      "name": "name-1",
		      "policyType": "DEFAULT_RESOURCE",
		      "groups": null
		    }
		  }
		}`

		response3 := `{
		  "data": {
		    "securityPolicy": null
		  }
		}`

		emptyResponses := []string{
			emptyResponse,
			response1,
			response2,
			response3,
		}

		c := newHTTPMockClient()

		for _, resp := range emptyResponses {
			httpmock.RegisterResponder("POST", c.GraphqlServerURL,
				MultipleResponders(
					httpmock.NewStringResponder(http.StatusOK, jsonResponse),
					httpmock.NewStringResponder(http.StatusOK, resp),
				),
			)

			securityPolicies, err := c.ReadSecurityPolicies(context.Background())

			httpmock.Reset()

			assert.Nil(t, securityPolicies)
			assert.EqualError(t, err, `failed to read security policy with id policy-1: failed to read group with id All: query result is empty`)
		}

	})
}
