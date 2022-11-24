package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateServiceAccountKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account Key - Ok", func(t *testing.T) {
		expected := &model.ServiceAccountKey{
			ID:               "key-id",
			Name:             "test",
			ServiceAccountID: "service-account-id",
			Status:           model.StatusActive,
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyCreate": {
		      "entity": {
		        "id": "key-id",
		        "name": "test",
		        "status": "ACTIVE",
		        "serviceAccount": {
		          "id": "service-account-id",
		          "name": "service-account-test"
		        }
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

		createServiceAccountKey, err := c.CreateServiceAccountKey(context.Background(), &model.ServiceAccountKey{
			ServiceAccountID: "service-account-id",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, expected, createServiceAccountKey)
	})
}

func TestCreateServiceAccountKeyWithNameOk(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account Key With Name - Ok", func(t *testing.T) {
		expected := &model.ServiceAccountKey{
			ID:               "key-id",
			Name:             "new name",
			ServiceAccountID: "service-account-id",
			Status:           model.StatusActive,
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyCreate": {
		      "entity": {
		        "id": "key-id",
		        "name": "new name",
		        "status": "ACTIVE",
		        "serviceAccount": {
		          "id": "service-account-id",
		          "name": "service-account-test"
		        }
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

		createServiceAccountKey, err := c.CreateServiceAccountKey(context.Background(), &model.ServiceAccountKey{
			ServiceAccountID: "service-account-id",
			Name:             "new name",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, expected, createServiceAccountKey)
	})
}

func TestCreateServiceAccountKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccount, err := c.CreateServiceAccountKey(context.Background(), &model.ServiceAccountKey{
			ServiceAccountID: "service-account-id",
		})

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, fmt.Sprintf(`failed to create service account key: Post "%s": bad request`, c.GraphqlServerURL))
	})
}

func TestCreateServiceAccountKeyResponseError(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account Key - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyCreate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccount, err := c.CreateServiceAccountKey(context.Background(), &model.ServiceAccountKey{
			ServiceAccountID: "service-account-id",
		})

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to create service account key: error_1`)
	})
}

func TestCreateServiceAccountKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccount, err := c.CreateServiceAccountKey(context.Background(), &model.ServiceAccountKey{})

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to create service account key: id is empty`)
	})
}

func TestCreateServiceAccountKeyWithNilRequest(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Account Key - With Nil Request", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccount, err := c.CreateServiceAccountKey(context.Background(), nil)

		assert.Nil(t, serviceAccount)
		assert.EqualError(t, err, `failed to create service account key: id is empty`)
	})
}

func TestReadServiceAccountKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account Key - Ok", func(t *testing.T) {
		expected := &model.ServiceAccountKey{
			ID:               "key-id",
			Name:             "name",
			ServiceAccountID: "service-account-id",
			Status:           model.StatusActive,
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccountKey": {
		      "id": "key-id",
		      "name": "name",
		      "status": "ACTIVE",
		      "serviceAccount": {
		        "id": "service-account-id",
		        "name": "service-account-test"
		      }
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccountKey, err := c.ReadServiceAccountKey(context.Background(), "key-id")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceAccountKey)
	})
}

func TestReadServiceAccountKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccountKey, err := c.ReadServiceAccountKey(context.Background(), "")

		assert.Nil(t, serviceAccountKey)
		assert.EqualError(t, err, `failed to read service account key: id is empty`)
	})
}

func TestReadServiceAccountKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccountKey, err := c.ReadServiceAccountKey(context.Background(), "account-key-id")

		assert.Nil(t, serviceAccountKey)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read service account key with id account-key-id: Post "%s": bad request`, c.GraphqlServerURL))
	})
}

func TestReadServiceAccountKeyEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Account Key - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKey": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceAccountKey, err := c.ReadServiceAccountKey(context.Background(), "account-key-id")

		assert.Nil(t, serviceAccountKey)
		assert.EqualError(t, err, `failed to read service account key with id account-key-id: query result is empty`)
	})
}

func TestUpdateServiceAccountKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Key - Ok", func(t *testing.T) {
		expected := &model.ServiceAccountKey{
			ID:               "key-id",
			Name:             "new name",
			ServiceAccountID: "service-account-id",
			Status:           model.StatusActive,
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyUpdate": {
		      "entity": {
		        "id": "key-id",
		        "name": "new name",
		        "status": "ACTIVE",
		        "serviceAccount": {
		          "id": "service-account-id",
		          "name": "service-account-test"
		        }
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

		serviceAccount, err := c.UpdateServiceAccountKey(context.Background(), &model.ServiceAccountKey{
			ID:   "key-id",
			Name: "new name",
		})

		assert.NoError(t, err)
		assert.Equal(t, expected, serviceAccount)
	})
}

func TestUpdateServiceAccountKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceAccountKey, err := c.UpdateServiceAccountKey(context.Background(), &model.ServiceAccountKey{})

		assert.Nil(t, serviceAccountKey)
		assert.EqualError(t, err, `failed to update service account key: id is empty`)
	})
}

func TestUpdateServiceAccountKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceAccountKey, err := c.UpdateServiceAccountKey(context.Background(), &model.ServiceAccountKey{
			ID:   "key-id",
			Name: "new name",
		})

		assert.Nil(t, serviceAccountKey)
		assert.EqualError(t, err, fmt.Sprintf(`failed to update service account key with id key-id: Post "%s": bad request`, c.GraphqlServerURL))
	})
}

func TestUpdateServiceAccountKeyResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Key - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyUpdate": {
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

		serviceAccountKey, err := c.UpdateServiceAccountKey(context.Background(), &model.ServiceAccountKey{
			ID:   "key-id",
			Name: "new name",
		})

		assert.Nil(t, serviceAccountKey)
		assert.EqualError(t, err, `failed to update service account key with id key-id: error_1`)
	})
}

func TestUpdateServiceAccountKeyEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Account Key - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyUpdate": {
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

		serviceAccountKey, err := c.UpdateServiceAccountKey(context.Background(), &model.ServiceAccountKey{
			ID:   "key-id",
			Name: "new name",
		})

		assert.Nil(t, serviceAccountKey)
		assert.EqualError(t, err, `failed to update service account key with id key-id: query result is empty`)
	})
}

func TestDeleteServiceAccountKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Account Key - Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyDelete": {
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := c.DeleteServiceAccountKey(context.Background(), "key-id")

		assert.NoError(t, err)
	})
}

func TestDeleteServiceAccountKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Account Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.DeleteServiceAccountKey(context.Background(), "")

		assert.EqualError(t, err, `failed to delete service account key: id is empty`)
	})
}

func TestDeleteServiceAccountKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Account Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := c.DeleteServiceAccountKey(context.Background(), "key-id")

		assert.EqualError(t, err, fmt.Sprintf(`failed to delete service account key with id key-id: Post "%s": bad request`, c.GraphqlServerURL))
	})
}

func TestDeleteServiceAccountKeyResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Account Key - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyDelete": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := c.DeleteServiceAccountKey(context.Background(), "key-id")

		assert.EqualError(t, err, `failed to delete service account key with id key-id: error_1`)
	})
}

func TestRevokeServiceAccountKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource : Revoke Service Account Key - Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyRevoke": {
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := c.RevokeServiceAccountKey(context.Background(), "key-id")

		assert.NoError(t, err)
	})
}

func TestRevokeServiceAccountKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Revoke Service Account Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.RevokeServiceAccountKey(context.Background(), "")

		assert.EqualError(t, err, `failed to revoke service account key: id is empty`)
	})
}

func TestRevokeServiceAccountKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Revoke Service Account Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := c.RevokeServiceAccountKey(context.Background(), "key-id")

		assert.EqualError(t, err, fmt.Sprintf(`failed to revoke service account key with id key-id: Post "%s": bad request`, c.GraphqlServerURL))
	})
}

func TestRevokeServiceAccountKeyResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Revoke Service Account Key - Response Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyRevoke": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		err := c.RevokeServiceAccountKey(context.Background(), "key-id")

		assert.EqualError(t, err, `failed to revoke service account key with id key-id: error_1`)
	})
}
