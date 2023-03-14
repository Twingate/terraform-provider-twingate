package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateServiceKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Key - Ok", func(t *testing.T) {
		expected := &model.ServiceKey{
			ID:      "key-id",
			Name:    "test",
			Service: "service-id",
			Status:  model.StatusActive,
			Token:   "token",
		}

		jsonResponse := `{
		  "data": {
		    "serviceAccountKeyCreate": {
		      "entity": {
		        "id": "key-id",
		        "name": "test",
		        "status": "ACTIVE",
		        "serviceAccount": {
		          "id": "service-id",
		          "name": "service-test"
		        }
		      },
		      "token": "token",
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceKey, err := c.CreateServiceKey(context.Background(), &model.ServiceKey{
			Service: "service-id",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceKey)
	})
}

func TestCreateServiceKeyWithNameOk(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Key With Name - Ok", func(t *testing.T) {
		expected := &model.ServiceKey{
			ID:      "key-id",
			Name:    "new name",
			Service: "service-account-id",
			Status:  model.StatusActive,
			Token:   "token",
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
		      "token": "token",
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceKey, err := c.CreateServiceKey(context.Background(), &model.ServiceKey{
			Service: "service-account-id",
			Name:    "new name",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceKey)
	})
}

func TestCreateServiceKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceKey, err := c.CreateServiceKey(context.Background(), &model.ServiceKey{
			Service: "service-account-id",
		})

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, graphqlErr(c, "failed to create service account key", errBadRequest))
	})
}

func TestCreateServiceKeyResponseError(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Key - Response Error", func(t *testing.T) {
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

		serviceKey, err := c.CreateServiceKey(context.Background(), &model.ServiceKey{
			Service: "service-account-id",
		})

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, `failed to create service account key: error_1`)
	})
}

func TestCreateServiceKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceKey, err := c.CreateServiceKey(context.Background(), &model.ServiceKey{})

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, `failed to create service account key: id is empty`)
	})
}

func TestCreateServiceKeyWithNilRequest(t *testing.T) {
	t.Run("Test Twingate Resource: Create Service Key - With Nil Request", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceKey, err := c.CreateServiceKey(context.Background(), nil)

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, `failed to create service account key: id is empty`)
	})
}

func TestReadServiceKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Key - Ok", func(t *testing.T) {
		expected := &model.ServiceKey{
			ID:      "key-id",
			Name:    "name",
			Service: "service-account-id",
			Status:  model.StatusActive,
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

		serviceKey, err := c.ReadServiceKey(context.Background(), "key-id")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, serviceKey)
	})
}

func TestReadServiceKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceKey, err := c.ReadServiceKey(context.Background(), "")

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, `failed to read service account key: id is empty`)
	})
}

func TestReadServiceKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceKey, err := c.ReadServiceKey(context.Background(), "account-key-id")

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, graphqlErr(c, "failed to read service account key with id account-key-id", errBadRequest))
	})
}

func TestReadServiceKeyEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource: Read Service Key - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "serviceAccountKey": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(http.StatusOK, jsonResponse))

		serviceKey, err := c.ReadServiceKey(context.Background(), "account-key-id")

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, `failed to read service account key with id account-key-id: query result is empty`)
	})
}

func TestUpdateServiceKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Key - Ok", func(t *testing.T) {
		expected := &model.ServiceKey{
			ID:      "key-id",
			Name:    "new name",
			Service: "service-account-id",
			Status:  model.StatusActive,
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

		serviceKey, err := c.UpdateServiceKey(context.Background(), &model.ServiceKey{
			ID:   "key-id",
			Name: "new name",
		})

		assert.NoError(t, err)
		assert.Equal(t, expected, serviceKey)
	})
}

func TestUpdateServiceKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		serviceKey, err := c.UpdateServiceKey(context.Background(), &model.ServiceKey{})

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, `failed to update service account key: id is empty`)
	})
}

func TestUpdateServiceKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		serviceKey, err := c.UpdateServiceKey(context.Background(), &model.ServiceKey{
			ID:   "key-id",
			Name: "new name",
		})

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, graphqlErr(c, "failed to update service account key with id key-id", errBadRequest))
	})
}

func TestUpdateServiceKeyResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Key - Response Error", func(t *testing.T) {
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

		serviceKey, err := c.UpdateServiceKey(context.Background(), &model.ServiceKey{
			ID:   "key-id",
			Name: "new name",
		})

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, `failed to update service account key with id key-id: error_1`)
	})
}

func TestUpdateServiceKeyEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Update Service Key - Empty Response", func(t *testing.T) {
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

		serviceKey, err := c.UpdateServiceKey(context.Background(), &model.ServiceKey{
			ID:   "key-id",
			Name: "new name",
		})

		assert.Nil(t, serviceKey)
		assert.EqualError(t, err, `failed to update service account key with id key-id: query result is empty`)
	})
}

func TestDeleteServiceKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Key - Ok", func(t *testing.T) {
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

		err := c.DeleteServiceKey(context.Background(), "key-id")

		assert.NoError(t, err)
	})
}

func TestDeleteServiceKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.DeleteServiceKey(context.Background(), "")

		assert.EqualError(t, err, `failed to delete service account key: id is empty`)
	})
}

func TestDeleteServiceKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := c.DeleteServiceKey(context.Background(), "key-id")

		assert.EqualError(t, err, graphqlErr(c, "failed to delete service account key with id key-id", errBadRequest))
	})
}

func TestDeleteServiceKeyResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Service Key - Response Error", func(t *testing.T) {
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

		err := c.DeleteServiceKey(context.Background(), "key-id")

		assert.EqualError(t, err, `failed to delete service account key with id key-id: error_1`)
	})
}

func TestRevokeServiceKeyOk(t *testing.T) {
	t.Run("Test Twingate Resource : Revoke Service Key - Ok", func(t *testing.T) {
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

		err := c.RevokeServiceKey(context.Background(), "key-id")

		assert.NoError(t, err)
	})
}

func TestRevokeServiceKeyWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Revoke Service Key - With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()

		err := c.RevokeServiceKey(context.Background(), "")

		assert.EqualError(t, err, `failed to revoke service account key: id is empty`)
	})
}

func TestRevokeServiceKeyRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Revoke Service Key - Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := c.RevokeServiceKey(context.Background(), "key-id")

		assert.EqualError(t, err, graphqlErr(c, "failed to revoke service account key with id key-id", errBadRequest))
	})
}

func TestRevokeServiceKeyResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Revoke Service Key - Response Error", func(t *testing.T) {
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

		err := c.RevokeServiceKey(context.Background(), "key-id")

		assert.EqualError(t, err, `failed to revoke service account key with id key-id: error_1`)
	})
}
