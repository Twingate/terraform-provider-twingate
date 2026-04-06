package client

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestReadSSHCertificateAuthorities(t *testing.T) {
	cases := []struct {
		name         string
		responseBody string
		expected     []*model.CertificateAuthority
		expectedErr  bool
	}{
		{
			name:         "empty edges - returns empty, no error",
			responseBody: `{"data":{"certificateAuthorities":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[]}}}`,
			expected:     []*model.CertificateAuthority{},
		},
		{
			name: "only SSH CAs - all returned",
			responseBody: `{"data":{"certificateAuthorities":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[
				{"node":{"__typename":"SSHCertificateAuthority","id":"ssh-ca-1","name":"ssh-ca-1","fingerprint":"fp1"}},
				{"node":{"__typename":"SSHCertificateAuthority","id":"ssh-ca-2","name":"ssh-ca-2","fingerprint":"fp2"}}
			]}}}`,
			expected: []*model.CertificateAuthority{
				{ID: "ssh-ca-1", Name: "ssh-ca-1", Fingerprint: "fp1"},
				{ID: "ssh-ca-2", Name: "ssh-ca-2", Fingerprint: "fp2"},
			},
		},
		{
			name: "mixed CA types - only SSH CAs returned",
			responseBody: `{"data":{"certificateAuthorities":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[
				{"node":{"__typename":"SSHCertificateAuthority","id":"ssh-ca-1","name":"ssh-ca","fingerprint":"fp1"}},
				{"node":{"__typename":"X509CertificateAuthority","id":"x509-ca-1","name":"x509-ca","fingerprint":"fp2"}}
			]}}}`,
			expected: []*model.CertificateAuthority{
				{ID: "ssh-ca-1", Name: "ssh-ca", Fingerprint: "fp1"},
			},
		},
		{
			name:         "graphql error - error propagated",
			responseBody: `{"errors":[{"message":"server error","locations":[{"line":1,"column":1}]}]}`,
			expectedErr:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := newTestClient()
			httpmock.ActivateNonDefault(client.HTTPClient)
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder("POST", client.GraphqlServerURL,
				httpmock.NewStringResponder(200, c.responseBody))

			authorities, err := client.ReadSSHCertificateAuthorities(context.Background())

			if c.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, authorities)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, authorities)
			}
		})
	}
}
