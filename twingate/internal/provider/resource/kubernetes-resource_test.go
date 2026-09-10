package resource

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/provider/providerdata"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testK8sID            = "k8s-resource-id"
	testK8sName          = "k8s-resource"
	testK8sGatewayID     = "gateway-id"
	testK8sNetworkID     = "remote-network-id"
	testK8sCustomAddress = "k8s-api.example.com"
	testK8sBearerToken   = "/var/run/secrets/token"
	testK8sCAFile        = "/var/run/secrets/ca.crt"
	testK8sAlias         = "k8s.internal"

	graphqlErrorResponse = `{"errors":[{"message":"api failure"}]}`
)

func kubernetesResourceSchema(t *testing.T) schema.Schema {
	t.Helper()

	resp := &resource.SchemaResponse{}
	(&kubernetesResource{}).Schema(t.Context(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	return resp.Schema
}

// inClusterKubernetesModel is the smallest valid plan for a Kubernetes Resource:
// it keeps the in_cluster default and omits the address so the resource has to
// fill in the fixed in-cluster address itself.
func inClusterKubernetesModel(t *testing.T) kubernetesResourceModel {
	t.Helper()

	return kubernetesResourceModel{
		ID:               types.StringUnknown(),
		Name:             types.StringValue(testK8sName),
		Address:          types.StringNull(),
		BearerTokenFile:  types.StringNull(),
		CAFile:           types.StringNull(),
		GatewayID:        types.StringValue(testK8sGatewayID),
		RemoteNetworkID:  types.StringValue(testK8sNetworkID),
		InCluster:        types.BoolValue(true),
		IsVisible:        types.BoolValue(true),
		Alias:            types.StringNull(),
		SecurityPolicyID: types.StringNull(),
		Tags:             types.MapNull(types.StringType),
		AccessPolicy:     makeObjectsSetNull(t.Context(), accessPolicyAttributeTypes()),
		GroupAccess:      makeObjectsSetNull(t.Context(), accessGroupAttributeTypes()),
	}
}

func kubernetesPlan(t *testing.T, model kubernetesResourceModel) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: kubernetesResourceSchema(t)}
	diags := plan.Set(t.Context(), model)
	require.False(t, diags.HasError(), diags)

	return plan
}

func kubernetesState(t *testing.T, model kubernetesResourceModel) tfsdk.State {
	t.Helper()

	state := tfsdk.State{Schema: kubernetesResourceSchema(t)}
	diags := state.Set(t.Context(), model)
	require.False(t, diags.HasError(), diags)

	return state
}

// emptyKubernetesState is the null state the framework hands to Create and ImportState.
func emptyKubernetesState(t *testing.T) tfsdk.State {
	t.Helper()

	k8sSchema := kubernetesResourceSchema(t)

	return tfsdk.State{
		Schema: k8sSchema,
		Raw:    tftypes.NewValue(k8sSchema.Type().TerraformType(t.Context()), nil),
	}
}

func readKubernetesState(t *testing.T, state tfsdk.State) kubernetesResourceModel {
	t.Helper()

	var model kubernetesResourceModel

	diags := state.Get(t.Context(), &model)
	require.False(t, diags.HasError(), diags)

	return model
}

// kubernetesEntityJSON is the entity returned by the create and update mutations.
func kubernetesEntityJSON(address, alias string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"address":{"value":%q},"remoteNetwork":{"id":%q},"gateway":{"id":%q},`+
		`"isVisible":true,"alias":%q,"securityPolicy":null,"tags":[],"approvalMode":"","accessPolicy":null}`,
		testK8sID, testK8sName, address, testK8sNetworkID, testK8sGatewayID, alias)
}

// kubernetesNodeJSON is the resource node returned by the read query.
func kubernetesNodeJSON(address string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"address":{"value":%q},"remoteNetwork":{"id":%q},"gateway":{"id":%q},`+
		`"isVisible":true,"alias":"","securityPolicy":null,"tags":[],"approvalMode":"","accessPolicy":null,`+
		`"access":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[]}}`,
		testK8sID, testK8sName, address, testK8sNetworkID, testK8sGatewayID)
}

// newMockedKubernetesResource returns a resource whose API client is served by
// httpmock with responseBody, plus a pointer that captures the last request body.
func newMockedKubernetesResource(t *testing.T, responseBody string) (*kubernetesResource, *string) {
	t.Helper()

	apiClient := client.NewClient(t.Context(), "https://test.twindev.com", "xxxx",
		time.Second, 0, client.DefaultAgent, "test", client.CacheOptions{})

	httpmock.ActivateNonDefault(apiClient.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)

	var requestBody string

	httpmock.RegisterResponder(http.MethodPost, apiClient.GraphqlServerURL, func(req *http.Request) (*http.Response, error) {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		requestBody = string(raw)

		return httpmock.NewStringResponse(http.StatusOK, responseBody), nil
	})

	return &kubernetesResource{client: apiClient}, &requestBody
}

func assertOperationError(t *testing.T, diags diag.Diagnostics, operation, detail string) {
	t.Helper()

	require.True(t, diags.HasError())
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, fmt.Sprintf("failed to %s %s", operation, TwingateKubernetesResource), diags.Errors()[0].Summary())
	assert.Contains(t, diags.Errors()[0].Detail(), detail)
}

func TestKubernetesResourceMetadata(t *testing.T) {
	resp := &resource.MetadataResponse{}
	NewKubernetesResourceResource().Metadata(t.Context(), resource.MetadataRequest{}, resp)

	assert.Equal(t, TwingateKubernetesResource, resp.TypeName)
}

func TestKubernetesResourceSchemaIsValid(t *testing.T) {
	diags := kubernetesResourceSchema(t).ValidateImplementation(t.Context())

	assert.False(t, diags.HasError(), diags)
}

func TestKubernetesResourceConfigure(t *testing.T) {
	apiClient := &client.Client{}

	cases := []struct {
		name         string
		providerData any
		expected     *client.Client
	}{
		{
			name:         "provider not configured yet - client stays unset",
			providerData: nil,
			expected:     nil,
		},
		{
			name:         "unexpected provider data - client stays unset",
			providerData: "not provider data",
			expected:     nil,
		},
		{
			name:         "provider data - client is taken from it",
			providerData: &providerdata.ProviderData{Client: apiClient},
			expected:     apiClient,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			k8sResource := &kubernetesResource{}
			k8sResource.Configure(t.Context(), resource.ConfigureRequest{ProviderData: c.providerData}, &resource.ConfigureResponse{})

			assert.Same(t, c.expected, k8sResource.client)
		})
	}
}

func TestKubernetesResourceImportState(t *testing.T) {
	resp := &resource.ImportStateResponse{State: emptyKubernetesState(t)}
	(&kubernetesResource{}).ImportState(t.Context(), resource.ImportStateRequest{ID: testK8sID}, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	var id types.String

	diags := resp.State.GetAttribute(t.Context(), path.Root(attr.ID), &id)
	require.False(t, diags.HasError(), diags)
	assert.Equal(t, types.StringValue(testK8sID), id)
}

// Plans that slipped past the schema validators (for example when in_cluster is
// only known at apply time) are rejected before the API is called, on both create
// and update.
func TestKubernetesResourceRejectsInvalidPlan(t *testing.T) {
	cases := []struct {
		name     string
		modify   func(m *kubernetesResourceModel)
		expected error
	}{
		{
			name: "in_cluster false without bearer_token_file",
			modify: func(m *kubernetesResourceModel) {
				m.InCluster = types.BoolValue(false)
				m.Address = types.StringValue(testK8sCustomAddress)
				m.CAFile = types.StringValue(testK8sCAFile)
			},
			expected: ErrBearerTokenFileEmpty,
		},
		{
			name: "in_cluster false without ca_file",
			modify: func(m *kubernetesResourceModel) {
				m.InCluster = types.BoolValue(false)
				m.Address = types.StringValue(testK8sCustomAddress)
				m.BearerTokenFile = types.StringValue(testK8sBearerToken)
			},
			expected: ErrCAFileEmpty,
		},
		{
			name: "in_cluster false without address",
			modify: func(m *kubernetesResourceModel) {
				m.InCluster = types.BoolValue(false)
				m.BearerTokenFile = types.StringValue(testK8sBearerToken)
				m.CAFile = types.StringValue(testK8sCAFile)
			},
			expected: ErrAddressEmpty,
		},
		{
			name: "in_cluster true with custom address",
			modify: func(m *kubernetesResourceModel) {
				m.Address = types.StringValue(testK8sCustomAddress)
			},
			expected: ErrAddressCannotBeModified,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			model := inClusterKubernetesModel(t)
			c.modify(&model)

			plan := kubernetesPlan(t, model)

			// No client is configured: every case has to fail before reaching the API.
			k8sResource := &kubernetesResource{}

			createResp := &resource.CreateResponse{}
			k8sResource.Create(t.Context(), resource.CreateRequest{Plan: plan}, createResp)
			assertOperationError(t, createResp.Diagnostics, operationCreate, c.expected.Error())

			updateResp := &resource.UpdateResponse{}
			k8sResource.Update(t.Context(), resource.UpdateRequest{Plan: plan}, updateResp)
			assertOperationError(t, updateResp.Diagnostics, operationUpdate, c.expected.Error())
		})
	}
}

// With in_cluster true and no address configured, create must send the fixed
// in-cluster address to the API and surface whatever the API answers.
func TestKubernetesResourceCreateSendsInClusterAddress(t *testing.T) {
	k8sResource, requestBody := newMockedKubernetesResource(t, graphqlErrorResponse)

	resp := &resource.CreateResponse{}
	k8sResource.Create(t.Context(), resource.CreateRequest{Plan: kubernetesPlan(t, inClusterKubernetesModel(t))}, resp)

	assert.Contains(t, *requestBody, defaultKubernetesAddress)
	assertOperationError(t, resp.Diagnostics, operationCreate, "api failure")
}

// A successful create stores the API entity: the unknown bearer_token_file and
// ca_file that Terraform plans for omitted computed attributes become null, while
// a configured alias is kept.
func TestKubernetesResourceCreateStoresAPIResponse(t *testing.T) {
	k8sResource, _ := newMockedKubernetesResource(t,
		`{"data":{"kubernetesResourceCreate":{"ok":true,"error":null,"entity":`+kubernetesEntityJSON(defaultKubernetesAddress, testK8sAlias)+`}}}`)

	planModel := inClusterKubernetesModel(t)
	planModel.BearerTokenFile = types.StringUnknown()
	planModel.CAFile = types.StringUnknown()
	planModel.Alias = types.StringValue(testK8sAlias)

	resp := &resource.CreateResponse{State: emptyKubernetesState(t)}
	k8sResource.Create(t.Context(), resource.CreateRequest{Plan: kubernetesPlan(t, planModel)}, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	state := readKubernetesState(t, resp.State)
	assert.Equal(t, types.StringValue(testK8sID), state.ID)
	assert.Equal(t, types.StringValue(testK8sName), state.Name)
	assert.Equal(t, types.StringValue(defaultKubernetesAddress), state.Address)
	assert.Equal(t, types.StringValue(testK8sGatewayID), state.GatewayID)
	assert.Equal(t, types.StringValue(testK8sNetworkID), state.RemoteNetworkID)
	assert.Equal(t, types.BoolValue(true), state.InCluster)
	assert.Equal(t, types.BoolValue(true), state.IsVisible)
	assert.Equal(t, types.StringNull(), state.BearerTokenFile)
	assert.Equal(t, types.StringNull(), state.CAFile)
	assert.Equal(t, types.StringValue(testK8sAlias), state.Alias)
}

// Switching in_cluster from false to true leaves the old custom address in state;
// the update must replace it with the fixed in-cluster address in the API call.
func TestKubernetesResourceUpdateSendsInClusterAddress(t *testing.T) {
	k8sResource, requestBody := newMockedKubernetesResource(t, graphqlErrorResponse)

	stateModel := inClusterKubernetesModel(t)
	stateModel.ID = types.StringValue(testK8sID)
	stateModel.InCluster = types.BoolValue(false)
	stateModel.Address = types.StringValue(testK8sCustomAddress)
	stateModel.BearerTokenFile = types.StringValue(testK8sBearerToken)
	stateModel.CAFile = types.StringValue(testK8sCAFile)

	resp := &resource.UpdateResponse{}
	k8sResource.Update(t.Context(), resource.UpdateRequest{
		Plan:  kubernetesPlan(t, inClusterKubernetesModel(t)),
		State: kubernetesState(t, stateModel),
	}, resp)

	assert.Contains(t, *requestBody, testK8sID)
	assert.Contains(t, *requestBody, defaultKubernetesAddress)
	assert.NotContains(t, *requestBody, testK8sCustomAddress)
	assertOperationError(t, resp.Diagnostics, operationUpdate, "api failure")
}

// After an import the state has no in_cluster value, so read derives it from
// whether the address is the fixed in-cluster one.
func TestKubernetesResourceReadDerivesInClusterFromAddress(t *testing.T) {
	cases := []struct {
		name      string
		address   string
		inCluster bool
	}{
		{name: "in-cluster address", address: defaultKubernetesAddress, inCluster: true},
		{name: "custom address", address: testK8sCustomAddress, inCluster: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			k8sResource, _ := newMockedKubernetesResource(t, `{"data":{"resource":`+kubernetesNodeJSON(c.address)+`}}`)

			stateModel := inClusterKubernetesModel(t)
			stateModel.ID = types.StringValue(testK8sID)
			stateModel.InCluster = types.BoolNull()

			resp := &resource.ReadResponse{State: emptyKubernetesState(t)}
			k8sResource.Read(t.Context(), resource.ReadRequest{State: kubernetesState(t, stateModel)}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			state := readKubernetesState(t, resp.State)
			assert.Equal(t, types.StringValue(c.address), state.Address)
			assert.Equal(t, types.BoolValue(c.inCluster), state.InCluster)
			assert.Equal(t, types.StringValue(testK8sGatewayID), state.GatewayID)
		})
	}
}

// A resource deleted outside Terraform is dropped from state instead of erroring,
// so the next apply recreates it.
func TestKubernetesResourceReadRemovesMissingResource(t *testing.T) {
	k8sResource, _ := newMockedKubernetesResource(t, `{"data":{"resource":null}}`)

	stateModel := inClusterKubernetesModel(t)
	stateModel.ID = types.StringValue(testK8sID)

	resp := &resource.ReadResponse{State: kubernetesState(t, stateModel)}
	k8sResource.Read(t.Context(), resource.ReadRequest{State: kubernetesState(t, stateModel)}, resp)

	assert.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	assert.True(t, resp.State.Raw.IsNull())
}

func TestKubernetesResourceDelete(t *testing.T) {
	cases := []struct {
		name          string
		response      string
		expectedError string
	}{
		{
			name:     "deleted - no diagnostics",
			response: `{"data":{"resourceDelete":{"ok":true,"error":null}}}`,
		},
		{
			name:          "rejected by the API - error reported",
			response:      `{"data":{"resourceDelete":{"ok":false,"error":"resource is in use"}}}`,
			expectedError: "resource is in use",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			k8sResource, requestBody := newMockedKubernetesResource(t, c.response)

			stateModel := inClusterKubernetesModel(t)
			stateModel.ID = types.StringValue(testK8sID)

			resp := &resource.DeleteResponse{}
			k8sResource.Delete(t.Context(), resource.DeleteRequest{State: kubernetesState(t, stateModel)}, resp)

			assert.Contains(t, *requestBody, testK8sID)

			if c.expectedError == "" {
				assert.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

				return
			}

			assertOperationError(t, resp.Diagnostics, operationDelete, c.expectedError)
		})
	}
}
