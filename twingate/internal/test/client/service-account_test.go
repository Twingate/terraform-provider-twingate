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
		assert.EqualError(t, err, fmt.Sprintf(`failed to create service account: Post "%s": bad request`, c.GraphqlServerURL))
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

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "id")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceAccount)
	})
}

func TestReadServiceAccountWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "")

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

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "account-id")

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read service account with id account-id: Post "%s": bad request`, c.GraphqlServerURL))
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

		serviceAccount, err := c.ReadServiceAccount(context.Background(), "account-id")

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
		assert.EqualError(t, err, fmt.Sprintf(`failed to update service account with id account-id: Post "%s": bad request`, c.GraphqlServerURL))
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

		assert.EqualError(t, err, fmt.Sprintf(`failed to delete service account with id account-id: Post "%s": bad request`, c.GraphqlServerURL))
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

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

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

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read service account with id All: Post "%s": bad request`, c.GraphqlServerURL))
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

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

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

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read service account with id All: Post "%s": bad request`, c.GraphqlServerURL))
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

		serviceAccounts, err := c.ReadServiceAccounts(context.Background())

		assert.Nil(t, serviceAccounts)
		assert.EqualError(t, err, `failed to read service account with id All: query result is empty`)
	})
}
