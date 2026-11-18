package resource

import (
	"encoding/json"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDNSProfileID   = "dns-profile-id"
	testDNSProfileName = "dns-profile-name"
)

func dnsProfileModel(allowedDomains, deniedDomains types.Object) dnsFilteringProfileModel {
	return dnsFilteringProfileModel{
		ID:                 types.StringValue(testDNSProfileID),
		Name:               types.StringValue(testDNSProfileName),
		Priority:           types.Float64Value(1),
		FallbackMethod:     types.StringValue(model.FallbackMethodStrict),
		Groups:             defaultEmptySet(),
		AllowedDomains:     allowedDomains,
		DeniedDomains:      deniedDomains,
		ContentCategories:  types.ObjectNull(contentCategoriesAttributeTypes()),
		SecurityCategories: types.ObjectNull(securityCategoriesAttributeTypes()),
		PrivacyCategories:  types.ObjectNull(privacyCategoriesAttributeTypes()),
	}
}

func domainsObject(isAuthoritative bool, domains ...string) types.Object {
	return convertDomainsToTerraform(domains, types.BoolValue(isAuthoritative))
}

// dnsProfileJSON renders a DNS filtering profile node as the API returns it.
func dnsProfileJSON(t *testing.T, allowedDomains, deniedDomains []string) string {
	t.Helper()

	node, err := json.Marshal(map[string]any{
		"id":             testDNSProfileID,
		"name":           testDNSProfileName,
		"priority":       1,
		"fallbackMethod": model.FallbackMethodStrict,
		"allowedDomains": allowedDomains,
		"deniedDomains":  deniedDomains,
		"groups": map[string]any{
			"pageInfo": map[string]any{"endCursor": "", "hasNextPage": false},
			"edges":    []any{},
		},
	})
	require.NoError(t, err)

	return string(node)
}

func requestDomains(t *testing.T, requestBody, name string) []string {
	t.Helper()

	raw, ok := requestVariable(t, requestBody, name).([]any)
	require.True(t, ok, "variable %q is not a list: %s", name, requestBody)

	domains := make([]string, 0, len(raw))
	for _, item := range raw {
		domains = append(domains, item.(string))
	}

	return domains
}

// A non-authoritative domain list must be merged with the same list already on
// the API, so domains managed outside Terraform survive the update, while an
// authoritative list replaces it. State keeps only the domains Terraform manages.
func TestDNSFilteringProfileUpdateMergesNonAuthoritativeDomains(t *testing.T) {
	const (
		planAllowed   = "plan-allowed.com"
		planDenied    = "plan-denied.com"
		originAllowed = "origin-allowed.com"
		originDenied  = "origin-denied.com"
	)

	cases := []struct {
		name                 string
		allowedAuthoritative bool
		deniedAuthoritative  bool
		expectedAllowed      []string
		expectedDenied       []string
	}{
		{
			name:                 "allowed domains non-authoritative",
			allowedAuthoritative: false,
			deniedAuthoritative:  true,
			expectedAllowed:      []string{planAllowed, originAllowed},
			expectedDenied:       []string{planDenied},
		},
		{
			name:                 "denied domains non-authoritative",
			allowedAuthoritative: true,
			deniedAuthoritative:  false,
			expectedAllowed:      []string{planAllowed},
			expectedDenied:       []string{planDenied, originDenied},
		},
		{
			name:                 "both non-authoritative",
			allowedAuthoritative: false,
			deniedAuthoritative:  false,
			expectedAllowed:      []string{planAllowed, originAllowed},
			expectedDenied:       []string{planDenied, originDenied},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			readResponse := `{"data":{"dnsFilteringProfile":` +
				dnsProfileJSON(t, []string{originAllowed}, []string{originDenied}) + `}}`
			updateResponse := `{"data":{"dnsFilteringProfileUpdate":{"ok":true,"error":null,"entity":` +
				dnsProfileJSON(t, c.expectedAllowed, c.expectedDenied) + `}}}`

			apiClient, requestBodies := newSequencedMockedClient(t, readResponse, updateResponse)
			profileResource := &dnsFilteringProfile{client: apiClient}

			planModel := dnsProfileModel(
				domainsObject(c.allowedAuthoritative, planAllowed),
				domainsObject(c.deniedAuthoritative, planDenied),
			)

			resp := &resource.UpdateResponse{State: nullStateOf(t, profileResource)}
			profileResource.Update(t.Context(), resource.UpdateRequest{
				Plan:  planOf(t, profileResource, planModel),
				State: stateOf(t, profileResource, planModel),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Len(t, *requestBodies, 2, "expected a read of the origin profile followed by the update")

			updateRequest := (*requestBodies)[1]
			assert.ElementsMatch(t, c.expectedAllowed, requestDomains(t, updateRequest, "allowedDomains"))
			assert.ElementsMatch(t, c.expectedDenied, requestDomains(t, updateRequest, "deniedDomains"))

			var state dnsFilteringProfileModel
			readStateInto(t, resp.State, &state)

			assert.ElementsMatch(t, []string{planAllowed}, convertDomains(state.AllowedDomains))
			assert.ElementsMatch(t, []string{planDenied}, convertDomains(state.DeniedDomains))
			assert.Equal(t, types.BoolValue(c.allowedAuthoritative), state.AllowedDomains.Attributes()[attr.IsAuthoritative])
			assert.Equal(t, types.BoolValue(c.deniedAuthoritative), state.DeniedDomains.Attributes()[attr.IsAuthoritative])
		})
	}
}

// When both lists are authoritative the origin profile is not read and the
// planned domains replace whatever the API holds.
func TestDNSFilteringProfileUpdateAuthoritativeDomainsSkipOriginRead(t *testing.T) {
	updateResponse := `{"data":{"dnsFilteringProfileUpdate":{"ok":true,"error":null,"entity":` +
		dnsProfileJSON(t, []string{"plan-allowed.com"}, []string{"plan-denied.com"}) + `}}}`

	apiClient, requestBodies := newSequencedMockedClient(t, updateResponse)
	profileResource := &dnsFilteringProfile{client: apiClient}

	planModel := dnsProfileModel(
		domainsObject(true, "plan-allowed.com"),
		domainsObject(true, "plan-denied.com"),
	)

	resp := &resource.UpdateResponse{State: nullStateOf(t, profileResource)}
	profileResource.Update(t.Context(), resource.UpdateRequest{
		Plan:  planOf(t, profileResource, planModel),
		State: stateOf(t, profileResource, planModel),
	}, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Len(t, *requestBodies, 1)
	assert.Equal(t, []string{"plan-allowed.com"}, requestDomains(t, (*requestBodies)[0], "allowedDomains"))
	assert.Equal(t, []string{"plan-denied.com"}, requestDomains(t, (*requestBodies)[0], "deniedDomains"))
}
