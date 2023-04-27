package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClientUserReadOk(t *testing.T) {
	testData := []struct {
		role    string
		isAdmin bool
	}{
		{role: "ADMIN", isAdmin: true},
		{role: "DEVOPS", isAdmin: false},
	}

	for _, td := range testData {
		t.Run("Test Twingate Resource : Read User Ok - "+td.role, func(t *testing.T) {
			const (
				userID = "id"
				email  = "user@email"
			)
			jsonResponse := fmt.Sprintf(`{
			  "data": {
			    "user": {
			      "id": "%s",
			      "firstName": "First",
			      "lastName": "Last",
			      "email": "%s",
			      "role": "%s"
			    }
			  }
			}`, userID, email, td.role)

			client := newHTTPMockClient()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", client.GraphqlServerURL,
				httpmock.NewStringResponder(200, jsonResponse))

			user, err := client.ReadUser(context.Background(), userID)

			assert.Nil(t, err)
			assert.NotNil(t, user)
			assert.EqualValues(t, userID, user.ID)
			assert.EqualValues(t, email, user.Email)
			assert.EqualValues(t, td.role, user.Role)
			assert.EqualValues(t, td.isAdmin, user.IsAdmin())
		})
	}
}

func TestClientUserReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read User Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "user": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const userID = "userID"
		user, err := client.ReadUser(context.Background(), userID)

		assert.Nil(t, user)
		assert.EqualError(t, err, fmt.Sprintf("failed to read user with id %s: query result is empty", userID))
	})
}

func TestClientUserReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read User Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		user, err := client.ReadUser(context.Background(), "userID")

		assert.Nil(t, user)
		assert.EqualError(t, err, graphqlErr(client, "failed to read user with id userID", errBadRequest))
	})
}

func TestClientReadEmptyUserError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty User Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "user": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		user, err := client.ReadUser(context.Background(), "")

		assert.EqualError(t, err, "failed to read user: id is empty")
		assert.Nil(t, user)
	})
}

func TestClientUserCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Create User Ok", func(t *testing.T) {
		input := &model.User{
			Email: "some@email.com",
			Role:  "MEMBER",
		}

		expected := &model.User{
			ID:    "user-id",
			Email: "some@email.com",
			Role:  "MEMBER",
		}

		jsonResponse := `{
          "data": {
            "userCreate": {
              "entity": {
                "id": "user-id",
                "email": "some@email.com",
                "role": "MEMBER"
              },
              "ok": true,
              "error": null
            }
          }
        }`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		user, err := client.CreateUser(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})
}

func TestClientUserCreateErrorEmptyEmail(t *testing.T) {
	t.Run("Test Twingate Resource : Create User - Error Empty Email", func(t *testing.T) {
		input := &model.User{
			Role: "MEMBER",
		}

		client := newHTTPMockClient()
		user, err := client.CreateUser(context.Background(), input)

		assert.Nil(t, user)
		assert.EqualError(t, err, "failed to create user: email is empty")
	})
}

func TestClientUserCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Create User - Request Error", func(t *testing.T) {
		input := &model.User{
			Email: "some@email.com",
			Role:  "MEMBER",
		}

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		user, err := client.CreateUser(context.Background(), input)

		assert.Nil(t, user)
		assert.EqualError(t, err, graphqlErr(client, "failed to create user with name some@email.com", errBadRequest))
	})
}

func TestClientUserUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update User Ok", func(t *testing.T) {
		input := &model.UserUpdate{
			ID:        "user-id",
			FirstName: optionalString("New name"),
		}

		expected := &model.User{
			ID:        "user-id",
			Email:     "some@email.com",
			FirstName: "New name",
			Role:      "MEMBER",
		}

		jsonResponse := `{
          "data": {
            "userDetailsUpdate": {
              "entity": {
                "id": "user-id",
                "email": "some@email.com",
                "firstName": "New name",
                "role": "MEMBER"
              },
              "ok": true,
              "error": null
            }
          }
        }`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		user, err := client.UpdateUser(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})
}

func TestClientUserUpdateWithRoleOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update User With Role Ok", func(t *testing.T) {
		input := &model.UserUpdate{
			ID:        "user-id",
			FirstName: optionalString("New name"),
			Role:      optionalString("SUPPORT"),
			IsActive:  optionalBool(false),
		}

		expected := &model.User{
			ID:        "user-id",
			Email:     "some@email.com",
			FirstName: "New name",
			Role:      "SUPPORT",
			IsActive:  false,
		}

		response1 := `{
          "data": {
            "userDetailsUpdate": {
              "entity": {
                "id": "user-id",
                "email": "some@email.com",
                "firstName": "New name",
                "role": "MEMBER",
                "state": "DISABLED"
              },
              "ok": true,
              "error": null
            }
          }
        }`

		response2 := `{
          "data": {
            "userRoleUpdate": {
              "entity": {
                "id": "user-id",
                "email": "some@email.com",
                "firstName": "New name",
                "role": "SUPPORT"
              },
              "ok": true,
              "error": null
            }
          }
        }`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, response1),
				httpmock.NewStringResponder(200, response2),
			),
		)

		user, err := client.UpdateUser(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})
}

func TestClientUserUpdateOnlyRoleOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update User Only Role Ok", func(t *testing.T) {
		input := &model.UserUpdate{
			ID:   "user-id",
			Role: optionalString("SUPPORT"),
		}

		expected := &model.User{
			ID:    "user-id",
			Email: "some@email.com",
			Role:  "SUPPORT",
		}

		response := `{
          "data": {
            "userRoleUpdate": {
              "entity": {
                "id": "user-id",
                "email": "some@email.com",
                "role": "SUPPORT"
              },
              "ok": true,
              "error": null
            }
          }
        }`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, response),
		)

		user, err := client.UpdateUser(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})
}

func TestClientUserNoUpdatesOk(t *testing.T) {
	t.Run("Test Twingate Resource : No User Updates Ok", func(t *testing.T) {
		input := &model.UserUpdate{
			ID: "user-id",
		}

		expected := &model.User{
			ID:       "user-id",
			Email:    "some@email.com",
			Role:     "SUPPORT",
			IsActive: true,
		}

		response := `{
          "data": {
            "user": {
              "id": "user-id",
              "email": "some@email.com",
              "role": "SUPPORT",
              "state": "PENDING"
            }
          }
        }`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, response),
		)

		user, err := client.UpdateUser(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})
}

func TestClientUserUpdateErrorEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update User - Error Empty ID", func(t *testing.T) {
		client := newHTTPMockClient()
		user, err := client.UpdateUser(context.Background(), &model.UserUpdate{})

		assert.Nil(t, user)
		assert.EqualError(t, err, "failed to update user: id is empty")
	})
}

func TestClientUserUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update User - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		user, err := client.UpdateUser(context.Background(), &model.UserUpdate{
			ID:        "user-id",
			FirstName: optionalString("Bob"),
		})

		assert.Nil(t, user)
		assert.EqualError(t, err, graphqlErr(client, "failed to update user with id user-id", errBadRequest))
	})
}

func TestClientUserUpdateWithEmptyRoleRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update User With Empty Role - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		user, err := client.UpdateUser(context.Background(), &model.UserUpdate{
			ID: "user-id",
		})

		assert.Nil(t, user)
		assert.EqualError(t, err, graphqlErr(client, "failed to read user with id user-id", errBadRequest))
	})
}

func TestClientUserUpdateRoleRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update User Role - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		user, err := client.UpdateUser(context.Background(), &model.UserUpdate{
			ID:   "user-id",
			Role: optionalString("DEVOPS"),
		})

		assert.Nil(t, user)
		assert.EqualError(t, err, graphqlErr(client, "failed to update user with id user-id", errBadRequest))
	})
}

func TestClientUserUpdateRoleErrorEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update User Role - Error Empty ID", func(t *testing.T) {
		client := newHTTPMockClient()
		user, err := client.UpdateUserRole(context.Background(), &model.UserUpdate{})

		assert.Nil(t, user)
		assert.EqualError(t, err, "failed to update user: id is empty")
	})
}

func TestClientUserDeleteErrorEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Delete User - Error Empty ID", func(t *testing.T) {
		client := newHTTPMockClient()
		err := client.DeleteUser(context.Background(), "")

		assert.EqualError(t, err, "failed to delete user: id is empty")
	})
}

func TestClientUserDeleteRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete User - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := client.DeleteUser(context.Background(), "user-id")

		assert.EqualError(t, err, graphqlErr(client, "failed to delete user with id user-id", errBadRequest))
	})
}

func TestClientUserDeleteOk(t *testing.T) {
	t.Run("Test Twingate Resource : Delete User Ok", func(t *testing.T) {
		response := `{
          "data": {
            "userDelete": {
              "ok": true,
              "error": null
            }
          }
        }`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, response),
		)

		err := client.DeleteUser(context.Background(), "user-id")

		assert.NoError(t, err)
	})
}

func TestClientUserDeleteResponseError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete User Response Error", func(t *testing.T) {
		response := `{
          "data": {
            "userDelete": {
              "ok": false,
              "error": "backend error"
            }
          }
        }`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, response),
		)

		err := client.DeleteUser(context.Background(), "user-id")

		assert.EqualError(t, err, `failed to delete user with id user-id: backend error`)
	})
}
