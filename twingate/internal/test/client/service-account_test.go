package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

var errBadRequest = errors.New("bad request")

func TestCreateServiceAccountOk(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account - Ok", func(t *testing.T) {
		expected := &model.ServiceAccount{
			ID:   "id",
			Name: "test",
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccountCreate": {
		      "entity": {
		        "id": "id",
		        "name": "test"
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.CreateServiceAccount(context.Background(), "test")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceAccount)
	})
}

func TestCreateServiceAccountRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccount, err := c.CreateServiceAccount(context.Background(), "test")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, graphqlErr(c, "failed to create service account", errBadRequest))
	})
}

func TestCreateServiceAccountResponseError(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountCreate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.CreateServiceAccount(context.Background(), "test")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to create service account: error_1`)
	})
}

func TestCreateServiceAccountWithEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account - With Empty Name", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccount, err := c.CreateServiceAccount(context.Background(), "")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to create service account: name is empty`)
	})
}

func TestReadServiceAccountOk(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account - Ok", func(t *testing.T) {
		expected := &model.ServiceAccount{
			ID:   "id",
			Name: "test",
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccount": {
		      "id": "id",
		      "name": "test"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.ReadShallowServiceAccount(context.Background(), "id")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceAccount)
	})
}

func TestReadServiceAccountWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccount, err := c.ReadShallowServiceAccount(context.Background(), "")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to read service account: id is empty`)
	})
}

func TestReadServiceAccountRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccount, err := c.ReadShallowServiceAccount(context.Background(), "account-id")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id account-id", errBadRequest))
	})
}

func TestReadServiceAccountEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccount": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.ReadShallowServiceAccount(context.Background(), "account-id")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to read service account with id account-id: query result is empty`)
	})
}

func TestUpdateServiceAccountOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account - Ok", func(t *testing.T) {
		expected := &model.ServiceAccount{
			ID:   "account-id",
			Name: "new name",
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccountUpdate": {
		      "entity": {
		        "id": "account-id",
		        "name": "new name"
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.UpdateServiceAccount(context.Background(), &model.ServiceAccount{
			ID:   "account-id",
			Name: "new name",
		})

		assert.NoError(t, err)
		assert.Equal(t, expected, serviceAccount)
	})
}

func TestUpdateServiceAccountWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccount, err := c.UpdateServiceAccount(context.Background(), &model.ServiceAccount{})

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to update service account: id is empty`)
	})
}

func TestUpdateServiceAccountWithEmptyName(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account - With Empty Name", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccount, err := c.UpdateServiceAccount(context.Background(), &model.ServiceAccount{
			ID: "account-id",
		})

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to update service account: name is empty`)
	})
}

func TestUpdateServiceAccountRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccount, err := c.UpdateServiceAccount(context.Background(), &model.ServiceAccount{
			ID:   "account-id",
			Name: "new name",
		})

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, graphqlErr(c, "failed to update service account with id account-id", errBadRequest))
	})
}

func TestUpdateServiceAccountResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountUpdate": {
		      "entity": null,
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.UpdateServiceAccount(context.Background(), &model.ServiceAccount{
			ID:   "account-id",
			Name: "new name",
		})

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to update service account with id account-id: error_1`)
	})
}

func TestUpdateServiceAccountEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountUpdate": {
		      "entity": null,
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.UpdateServiceAccount(context.Background(), &model.ServiceAccount{
			ID:   "account-id",
			Name: "new name",
		})

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to update service account with id account-id: query result is empty`)
	})
}

func TestDeleteServiceAccountOk(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Account - Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountDelete": {
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := c.DeleteServiceAccount(context.Background(), "account-id")

		assert.NoError(t, err)
	})
}

func TestDeleteServiceAccountWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Account - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.DeleteServiceAccount(context.Background(), "")

		assert.EqualError(t, err, `failed to delete service account: id is empty`)
	})
}

func TestDeleteServiceAccountRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Account - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := c.DeleteServiceAccount(context.Background(), "account-id")

		assert.EqualError(t, err, graphqlErr(c, "failed to delete service account with id account-id", errBadRequest))
	})
}

func TestDeleteServiceAccountResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Account - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountDelete": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := c.DeleteServiceAccount(context.Background(), "account-id")

		assert.EqualError(t, err, `failed to delete service account with id account-id: error_1`)
	})
}

func TestReadServiceAccountsOk(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Accounts - Ok", func(t *testing.T) {
		expected := []*model.ServiceAccount{
			{
				ID:   "id-1",
				Name: "test-1",
			},
			{
				ID:   "id-2",
				Name: "test-2",
			},
			{
				ID:   "id-3",
				Name: "test-3",
			},
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1"
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "test-2"
		          }
		        }
		      ]
		    }
		  }
		}`

		nextResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-3",
		            "name": "test-3"
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
				httpmock.NewStringResponder(http.StatusOK, nextResponse),
			),
		)

		serviceAccounts, err := c.ReadShallowServiceAccounts(context.Background())

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceAccounts)
	})
}

func TestReadServiceAccountsRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Accounts - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccounts, err := c.ReadShallowServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id All", errBadRequest))
	})
}

func TestReadServiceAccountsEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Accounts - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "edges": []
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccounts, err := c.ReadShallowServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, `failed to read service account with id All: query result is empty`)
	})
}

func TestReadServiceAccountsRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Accounts - Request Error on Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1"
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

		serviceAccounts, err := c.ReadShallowServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id All", errBadRequest))
	})
}

func TestReadServiceAccountsEmptyResponseOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Accounts - Empty Response on Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1"
		          }
		        }
		      ]
		    }
		  }
		}`

		nextResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "hasNextPage": false
		      },
		      "edges": []
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

		serviceAccounts, err := c.ReadShallowServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, `failed to read service account with id All: query result is empty`)
	})
}

func TestReadServicesOk(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Ok", func(t *testing.T) {
		expected := []*model.ServiceAccount{
			{
				ID:        "id-1",
				Name:      "test-1",
				Resources: []string{},
				Keys:      []string{"key-1"},
			},
			{
				ID:        "id-2",
				Name:      "test-2",
				Resources: []string{"resource-2-1", "resource-2-2"},
				Keys:      []string{"key-2-1", "key-2-2"},
			},
			{
				ID:        "id-3",
				Name:      "test-3",
				Resources: []string{},
				Keys:      []string{"key-3-1", "key-3-2", "key-3-3"},
			},
		}

		response1 := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": null
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-01",
		                "hasNextPage": false
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-1",
		                    "status": "ACTIVE"
		                  }
		                }
		              ]
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "test-2",
		            "resources": {
		              "pageInfo": {
		                "endCursor": "cursor-resource-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "resource-2-1",
		                    "isActive": true
		                  }
		                }
		              ]
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-2-1",
		                    "status": "ACTIVE"
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
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-2",
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-3",
		            "name": "test-3",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": null
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-03",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-3-1",
		                    "status": "ACTIVE"
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
		    "serviceAccount": {
		      "id": "id-2",
		      "name": "test-2",
		      "resources": {
		        "pageInfo": {
		          "endCursor": "cursor-resource-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "resource-2-2",
		              "isActive": true
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		response4 := `{
		  "data": {
		    "serviceAccount": {
		      "id": "id-2",
		      "name": "test-2",
		      "keys": {
		        "pageInfo": {
		          "endCursor": "cursor-key-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "key-2-2",
		              "status": "ACTIVE"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		response5 := `{
		  "data": {
		    "serviceAccount": {
		      "id": "id-3",
		      "name": "test-3",
		      "resources": {
		        "pageInfo": {
		          "endCursor": null,
		          "hasNextPage": false
		        },
		        "edges": null
		      },
		      "keys": {
		        "pageInfo": {
		          "endCursor": "cursor-key-03-1",
		          "hasNextPage": true
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "key-3-2",
		              "status": "ACTIVE"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		response6 := `{
		  "data": {
		    "serviceAccount": {
		      "id": "id-3",
		      "name": "test-3",
		      "resources": {
		        "pageInfo": {
		          "endCursor": null,
		          "hasNextPage": false
		        },
		        "edges": null
		      },
		      "keys": {
		        "pageInfo": {
		          "endCursor": "cursor-key-03-2",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "key-3-3",
		              "status": "ACTIVE"
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
				httpmock.NewStringResponder(http.StatusOK, response5),
				httpmock.NewStringResponder(http.StatusOK, response6),
			),
		)

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceAccounts)
	})
}

func TestReadServicesRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id All", errBadRequest))
	})
}

func TestReadServicesEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "edges": []
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.NoError(t, err)
	})
}

func TestReadServicesRequestErrorOnFetchingServices(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Request Error on Fetching Services", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1"
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

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id All", errBadRequest))
	})
}

func TestReadServicesEmptyResponseOnFetchingServices(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Empty Response on Fetching Services", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1"
		          }
		        }
		      ]
		    }
		  }
		}`

		nextResponse := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "hasNextPage": false
		      },
		      "edges": []
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

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, `failed to read service account with id All: query result is empty`)
	})
}

func TestReadServicesRequestErrorOnFetchingResources(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Request Error on Fetching Resources", func(t *testing.T) {

		response1 := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": []
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-01",
		                "hasNextPage": false
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-1"
		                  }
		                }
		              ]
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "test-2",
		            "resources": {
		              "pageInfo": {
		                "endCursor": "cursor-resource-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "resource-2-1"
		                  }
		                }
		              ]
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-2-1"
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
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-2",
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-3",
		            "name": "test-3",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": []
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-03",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-3-1"
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
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id All", errBadRequest))
	})
}

func TestReadServicesEmptyResponseOnFetchingResources(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Empty Response on Fetching Resources", func(t *testing.T) {

		response1 := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": []
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-01",
		                "hasNextPage": false
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-1"
		                  }
		                }
		              ]
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "test-2",
		            "resources": {
		              "pageInfo": {
		                "endCursor": "cursor-resource-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "resource-2-1"
		                  }
		                }
		              ]
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-2-1"
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
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-2",
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-3",
		            "name": "test-3",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": []
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-03",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-3-1"
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
		    "serviceAccount": {
		      "id": "id-2",
		      "name": "test-2",
		      "resources": {
		        "pageInfo": {
		          "hasNextPage": false
		        },
		        "edges": []
		      },
		      "keys": {
		        "pageInfo": {
		          "hasNextPage": false
		        },
		        "edges": []
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
			),
		)

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, `failed to read service account with id All: query result is empty`)
	})
}

func TestReadServicesRequestErrorOnFetchingKeys(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Request Error on Fetching Keys", func(t *testing.T) {

		response1 := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": []
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-01",
		                "hasNextPage": false
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-1"
		                  }
		                }
		              ]
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "test-2",
		            "resources": {
		              "pageInfo": {
		                "endCursor": "cursor-resource-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "resource-2-1"
		                  }
		                }
		              ]
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-2-1"
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
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-2",
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-3",
		            "name": "test-3",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": []
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-03",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-3-1"
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
		    "serviceAccount": {
		      "id": "id-2",
		      "name": "test-2",
		      "resources": {
		        "pageInfo": {
		          "endCursor": "cursor-resource-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "resource-2-2"
		            }
		          }
		        ]
		      },
		      "keys": {
		        "pageInfo": {
		          "endCursor": "cursor-key-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "key-2-2"
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
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id All", errBadRequest))
	})
}

func TestReadServicesEmptyResponseOnFetchingKeys(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - Empty Response on Fetching Keys", func(t *testing.T) {

		response1 := `{
		  "data": {
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-1",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-1",
		            "name": "test-1",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": []
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-01",
		                "hasNextPage": false
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-1"
		                  }
		                }
		              ]
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "id-2",
		            "name": "test-2",
		            "resources": {
		              "pageInfo": {
		                "endCursor": "cursor-resource-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "resource-2-1"
		                  }
		                }
		              ]
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-02",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-2-1"
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
		    "serviceAccounts": {
		      "pageInfo": {
		        "endCursor": "cursor-2",
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "id-3",
		            "name": "test-3",
		            "resources": {
		              "pageInfo": {
		                "endCursor": null,
		                "hasNextPage": false
		              },
		              "edges": []
		            },
		            "keys": {
		              "pageInfo": {
		                "endCursor": "cursor-key-03",
		                "hasNextPage": true
		              },
		              "edges": [
		                {
		                  "node": {
		                    "id": "key-3-1"
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
		    "serviceAccount": {
		      "id": "id-2",
		      "name": "test-2",
		      "resources": {
		        "pageInfo": {
		          "endCursor": "cursor-resource-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "resource-2-2"
		            }
		          }
		        ]
		      },
		      "keys": {
		        "pageInfo": {
		          "endCursor": "cursor-key-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "key-2-2"
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		response4 := `{
		  "data": {
		    "serviceAccount": {
		      "id": "id-2",
		      "name": "test-2",
		      "resources": {
		        "pageInfo": {
		          "hasNextPage": false
		        },
		        "edges": []
		      },
		      "keys": {
		        "pageInfo": {
		          "hasNextPage": false
		        },
		        "edges": []
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

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, `failed to read service account with id All: query result is empty`)
	})
}

func TestReadServicesByNameOk(t *testing.T) {
	t.Run("Test Twingate Resource: Read Services - By Name Ok", func(t *testing.T) {
		expected := []*model.ServiceAccount{
			{
				ID:        "id-2",
				Name:      "test-2",
				Resources: []string{"resource-2-2"},
				Keys:      []string{"key-2-1"},
			},
		}

		response1 := `{
		 "data": {
		   "serviceAccounts": {
		     "pageInfo": {
		       "endCursor": "cursor-1",
		       "hasNextPage": false
		     },
		     "edges": [
		       {
		         "node": {
		           "id": "id-2",
		           "name": "test-2",
		           "resources": {
		             "pageInfo": {
		               "endCursor": "cursor-resource-02",
		               "hasNextPage": true
		             },
		             "edges": [
		               {
		                 "node": {
		                   "id": "resource-2-1",
		                   "isActive": false
		                 }
		               }
		             ]
		           },
		           "keys": {
		             "pageInfo": {
		               "endCursor": "cursor-key-02",
		               "hasNextPage": true
		             },
		             "edges": [
		               {
		                 "node": {
		                   "id": "key-2-1",
		                   "status": "ACTIVE"
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
		    "serviceAccount": {
		      "id": "id-2",
		      "name": "test-2",
		      "resources": {
		        "pageInfo": {
		          "endCursor": "cursor-resource-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "resource-2-2",
		              "isActive": true
		            }
		          }
		        ]
		      }
		    }
		  }
		}`

		response3 := `{
		  "data": {
		    "serviceAccount": {
		      "id": "id-2",
		      "name": "test-2",
		      "keys": {
		        "pageInfo": {
		          "endCursor": "cursor-key-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "key-2-2",
		              "status": "REVOKED"
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
				httpmock.NewStringResponder(http.StatusOK, response3),
			),
		)

		serviceAccounts, err := c.ReadServiceAccounts(context.Background(), "test-2")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceAccounts)
	})
}

func TestReadServiceOk(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service - Ok", func(t *testing.T) {
		expected := &model.ServiceAccount{
			ID:        "id",
			Name:      "test-2",
			Resources: []string{},
			Keys:      []string{},
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccount": {
		      "id": "id",
		      "name": "test-2",
		      "keys": {
		        "pageInfo": {
		          "endCursor": "cursor-key-02-1",
		          "hasNextPage": false
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "key-2-2",
		              "status": "REVOKED"
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
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "id")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceAccount)
	})
}

func TestReadServiceWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to read service account: id is empty`)
	})
}

func TestReadServiceRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "account-id")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id account-id", errBadRequest))
	})
}

func TestReadServiceEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccount": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "account-id")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to read service account with id account-id: query result is empty`)
	})
}

func TestReadServiceRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service - Request Error On Fetching", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccount": {
		      "id": "account-id",
		      "name": "test-2",
		      "keys": {
		        "pageInfo": {
		          "endCursor": "cursor-key-02-1",
		          "hasNextPage": true
		        },
		        "edges": [
		          {
		            "node": {
		              "id": "key-2-2",
		              "status": "REVOKED"
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
				httpmock.NewErrorResponder(errBadRequest),
			))

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "account-id")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account with id All", errBadRequest))
	})
}

func TestUpdateServiceAccountRemoveResourcesOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Remove Resources - Ok", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "serviceAccount": {
		      "id": "account-id",
		      "name": "test-1",
		      "keys": null
		      }
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "serviceAccountUpdate": {
		      "entity": {
		        "id": "account-id",
		        "name": "account name"
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
			))

		err := c.UpdateServiceAccountRemoveResources(context.Background(), "account-id", []string{"resource-1"})

		assert.NoError(t, err)
	})
}

func TestUpdateServiceAccountRemoveResourcesOkButEmpty(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Remove Resources - Ok But Empty", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "serviceAccount": {
		      "id": "account-id",
		      "name": "test-1",
		      "keys": null
		      }
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "serviceAccountUpdate": {
		      "entity": null,
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
			))

		err := c.UpdateServiceAccountRemoveResources(context.Background(), "account-id", []string{"resource-1"})

		assert.EqualError(t, err, `failed to update service account with id account-id: query result is empty`)
	})
}

func TestUpdateServiceAccountRemoveResourcesWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Remove Resources - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.UpdateServiceAccountRemoveResources(context.Background(), "", []string{"resource-1"})

		assert.EqualError(t, err, `failed to update service account: id is empty`)
	})
}

func TestUpdateServiceAccountRemoveResourcesWithEmptyResourceIDs(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Remove Resources - With Empty Resource IDs", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.UpdateServiceAccountRemoveResources(context.Background(), "service-id", nil)

		assert.NoError(t, err)
	})
}

func TestUpdateServiceAccountRemoveResourcesRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Update Service Account Remove Resources - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := c.UpdateServiceAccountRemoveResources(context.Background(), "service-id", []string{"id1"})

		assert.EqualError(t, err, graphqlErr(c, "failed to update service account with id service-id", errBadRequest))
	})
}

func TestUpdateServiceAccountRemoveResourcesResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Remove Resources - Response Error", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "serviceAccount": {
		      "id": "service-id",
		      "name": "test-1",
		      "keys": null
		      }
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "serviceAccountUpdate": {
		      "entity": null,
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
			))

		err := c.UpdateServiceAccountRemoveResources(context.Background(), "service-id", []string{"id1"})

		assert.EqualError(t, err, `failed to update service account with id service-id: error_1`)
	})
}

func TestUpdateServiceAccountRemoveResourcesEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Remove Resources - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccount": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := c.UpdateServiceAccountRemoveResources(context.Background(), "service-id", []string{"id1"})

		assert.NoError(t, err)
	})
}
