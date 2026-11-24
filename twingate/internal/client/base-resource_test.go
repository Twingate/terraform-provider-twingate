package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// kubernetesEntity is the mutation payload the API returns for a Kubernetes resource.
	kubernetesEntity = `{
	"id": "k8s-1",
	"name": "k8s-resource",
	"address": {"value": "kubernetes.default.svc.cluster.local"},
	"remoteNetwork": {"id": "rn-1"},
	"gateway": {"id": "gw-1"},
	"isVisible": true,
	"alias": "",
	"securityPolicy": null,
	"tags": [],
	"approvalMode": "",
	"accessPolicy": null
}`

	testAccessGroupID          = "R3JvdXA6MQ=="
	testAccessServiceAccountID = "U2VydmljZUFjY291bnQ6MQ=="
)

type pamClientCase struct {
	name     string
	mutation string
	entity   string
	create   func(client *Client, access model.PAMResource) (model.PAMResource, error)
	read     func(client *Client) error
}

func pamClientCases() []pamClientCase {
	return []pamClientCase{
		{
			name:     "kubernetes",
			mutation: "kubernetesResourceCreate",
			entity:   kubernetesEntity,
			create: func(client *Client, access model.PAMResource) (model.PAMResource, error) {
				res, err := client.CreateKubernetesResource(context.Background(), &model.KubernetesResource{PAMResource: access})
				if err != nil {
					return model.PAMResource{}, err
				}

				return res.PAMResource, nil
			},
			read: func(client *Client) error {
				_, err := client.ReadKubernetesResource(context.Background(), "k8s-1")

				return err
			},
		},
		{
			name:     "ssh",
			mutation: "sshResourceCreate",
			entity:   sshEntity,
			create: func(client *Client, access model.PAMResource) (model.PAMResource, error) {
				res, err := client.CreateSSHResource(context.Background(), &model.SSHResource{PAMResource: access})
				if err != nil {
					return model.PAMResource{}, err
				}

				return res.PAMResource, nil
			},
			read: func(client *Client) error {
				_, err := client.ReadSSHResource(context.Background(), "ssh-1")

				return err
			},
		},
		{
			name:     "web app",
			mutation: "webAppResourceCreate",
			entity:   webAppEntity,
			create: func(client *Client, access model.PAMResource) (model.PAMResource, error) {
				input := webAppInput()
				input.GroupsAccess = access.GroupsAccess
				input.ServiceAccounts = access.ServiceAccounts

				res, err := client.CreateWebAppResource(context.Background(), input)
				if err != nil {
					return model.PAMResource{}, err
				}

				return res.PAMResource, nil
			},
			read: func(client *Client) error {
				_, err := client.ReadWebAppResource(context.Background(), "web-app-1")

				return err
			},
		},
	}
}

// sequencedResponder answers successive requests with successive bodies and
// counts the requests it served.
func sequencedResponder(t *testing.T, requests *int, bodies ...string) httpmock.Responder {
	t.Helper()

	return func(req *http.Request) (*http.Response, error) {
		_, err := io.Copy(io.Discard, req.Body)
		require.NoError(t, err)

		*requests++

		if *requests > len(bodies) {
			return httpmock.NewStringResponse(http.StatusOK, bodies[len(bodies)-1]), nil
		}

		return httpmock.NewStringResponse(http.StatusOK, bodies[*requests-1]), nil
	}
}

// The planned access is assigned right after a PAM resource is created and is
// recorded on the returned model; a failure to assign it fails the create.
func TestCreatePAMResourceAssignsAccess(t *testing.T) {
	access := model.PAMResource{
		BaseResource: model.BaseResource{
			GroupsAccess:    []model.AccessGroup{{GroupID: testAccessGroupID}},
			ServiceAccounts: []string{testAccessServiceAccountID},
		},
	}

	for _, c := range pamClientCases() {
		created := `{"data":{"` + c.mutation + `":{"ok":true,"error":null,"entity":` + c.entity + `}}}`

		t.Run(c.name+": access assigned", func(t *testing.T) {
			client := newTestClient(t.Context())
			httpmock.ActivateNonDefault(client.HTTPClient)
			defer httpmock.DeactivateAndReset()

			var requests int

			httpmock.RegisterResponder("POST", client.GraphqlServerURL,
				sequencedResponder(t, &requests, created, `{"data":{"resourceAccessAdd":{"ok":true,"error":null}}}`))

			res, err := c.create(client, access)

			require.NoError(t, err)
			assert.Equal(t, 2, requests)
			assert.Equal(t, access.GroupsAccess, res.GroupsAccess)
			assert.Equal(t, access.ServiceAccounts, res.ServiceAccounts)
		})

		t.Run(c.name+": access failure fails the create", func(t *testing.T) {
			client := newTestClient(t.Context())
			httpmock.ActivateNonDefault(client.HTTPClient)
			defer httpmock.DeactivateAndReset()

			var requests int

			httpmock.RegisterResponder("POST", client.GraphqlServerURL,
				sequencedResponder(t, &requests, created, `{"data":{"resourceAccessAdd":{"ok":false,"error":"access denied"}}}`))

			_, err := c.create(client, access)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "access denied")
		})
	}
}

// An access edge without a principal ID makes the read fail instead of
// silently dropping the assignment.
func TestReadPAMResourceInvalidAccess(t *testing.T) {
	cases := []struct {
		typename string
		expected error
	}{
		{typename: query.AccessGroup, expected: query.ErrMissingAccessGroupID},
		{typename: query.AccessServiceAccount, expected: query.ErrMissingAccessServiceAccountID},
	}

	for _, pam := range pamClientCases() {
		for _, c := range cases {
			t.Run(pam.name+": "+c.typename, func(t *testing.T) {
				client := newTestClient(t.Context())
				httpmock.ActivateNonDefault(client.HTTPClient)
				defer httpmock.DeactivateAndReset()

				entity := strings.TrimSuffix(strings.TrimSpace(pam.entity), "}")
				body := `{"data":{"resource":` + entity + `,"access":{"pageInfo":{"endCursor":"","hasNextPage":false},` +
					`"edges":[{"node":{"__typename":"` + c.typename + `","id":""}}]}}}}`

				httpmock.RegisterResponder("POST", client.GraphqlServerURL, httpmock.NewStringResponder(200, body))

				err := pam.read(client)

				require.ErrorIs(t, err, c.expected)
			})
		}
	}
}
