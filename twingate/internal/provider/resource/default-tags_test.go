package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/provider/providerdata"
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

// The three gateway-backed resources inherit provider default tags the same way
// twingate_resource does: `tags` holds only what the user declared, `tags_all`
// is the merge of provider defaults and user tags and is what reaches the API.

// gatewayResourceCase describes one resource type under test. The GraphQL
// payloads mirror the shapes returned by the read query and create mutation.
type gatewayResourceCase struct {
	name        string
	newResource func() resource.Resource
	readQuery   string
	createQuery string
	// requiredAttrs are the attributes a valid plan/state must carry.
	requiredAttrs map[string]tftypes.Value
	// entityExtra holds resource-specific fields appended to API entity JSON.
	entityExtra string
}

func gatewayResourceCases() []gatewayResourceCase {
	return []gatewayResourceCase{
		{
			name:        TwingateSSHResource,
			newResource: NewSSHResourceResource,
			readQuery:   "resource",
			createQuery: "sshResourceCreate",
			requiredAttrs: map[string]tftypes.Value{
				attr.Name:            tftypes.NewValue(tftypes.String, "ssh"),
				attr.Address:         tftypes.NewValue(tftypes.String, "10.0.0.1"),
				attr.GatewayID:       tftypes.NewValue(tftypes.String, "gw-1"),
				attr.RemoteNetworkID: tftypes.NewValue(tftypes.String, "rn-1"),
			},
		},
		{
			name:        TwingateKubernetesResource,
			newResource: NewKubernetesResourceResource,
			readQuery:   "resource",
			createQuery: "kubernetesResourceCreate",
			requiredAttrs: map[string]tftypes.Value{
				attr.Name:            tftypes.NewValue(tftypes.String, "k8s"),
				attr.Address:         tftypes.NewValue(tftypes.String, defaultKubernetesAddress),
				attr.GatewayID:       tftypes.NewValue(tftypes.String, "gw-1"),
				attr.RemoteNetworkID: tftypes.NewValue(tftypes.String, "rn-1"),
				attr.InCluster:       tftypes.NewValue(tftypes.Bool, true),
			},
		},
		{
			name:        TwingateWebAppResource,
			newResource: NewWebAppResourceResource,
			readQuery:   "resource",
			createQuery: "webAppResourceCreate",
			requiredAttrs: map[string]tftypes.Value{
				attr.Name:            tftypes.NewValue(tftypes.String, "web"),
				attr.Address:         tftypes.NewValue(tftypes.String, "internal.acme.com"),
				attr.GatewayID:       tftypes.NewValue(tftypes.String, "gw-1"),
				attr.RemoteNetworkID: tftypes.NewValue(tftypes.String, "rn-1"),
				attr.Upstream:        portObject(8080),
				attr.Downstream:      portObject(80),
			},
			entityExtra: `,"upstream":{"port":8080},"downstream":{"port":80},"requestHeaderRewrites":[]`,
		},
	}
}

func portObject(port int64) tftypes.Value {
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{attr.Port: tftypes.Number}}

	return tftypes.NewValue(objType, map[string]tftypes.Value{
		attr.Port: tftypes.NewValue(tftypes.Number, port),
	})
}

func tagsJSON(tags map[string]string) string {
	items := make([]map[string]string, 0, len(tags))
	for k, v := range tags {
		items = append(items, map[string]string{"key": k, "value": v})
	}

	raw, _ := json.Marshal(items)

	return string(raw)
}

// entityJSON renders the shared fields of a resource node. The read query also
// selects the paginated `access` connection, the create mutation does not, and
// the GraphQL client rejects fields that are absent from the selection.
func (c gatewayResourceCase) entityJSON(id string, tags map[string]string, withAccess bool) string {
	access := ""
	if withAccess {
		access = `,"access":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[]}`
	}

	return fmt.Sprintf(`{"id":%q,"name":"res","address":{"value":"10.0.0.1"},"remoteNetwork":{"id":"rn-1"},"gateway":{"id":"gw-1"},`+
		`"isVisible":true,"alias":null,"securityPolicy":null,"tags":%s,"approvalMode":"MANUAL","accessPolicy":null%s%s}`,
		id, tagsJSON(tags), access, c.entityExtra)
}

func (c gatewayResourceCase) readResponse(id string, tags map[string]string) string {
	return fmt.Sprintf(`{"data":{%q:%s}}`, c.readQuery, c.entityJSON(id, tags, true))
}

func (c gatewayResourceCase) createResponse(id string, tags map[string]string) string {
	return fmt.Sprintf(`{"data":{%q:{"ok":true,"error":null,"entity":%s}}}`, c.createQuery, c.entityJSON(id, tags, false))
}

// mockedClient returns a client whose HTTP transport is intercepted by httpmock.
// Every request body is appended to the returned slice for later inspection.
func mockedClient(t *testing.T) (*client.Client, *[]map[string]any) {
	t.Helper()

	c := client.NewClient(t.Context(), "https://test.twindev.com", "token", time.Second, 0, client.DefaultAgent, "test", client.CacheOptions{})
	httpmock.ActivateNonDefault(c.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)

	requests := &[]map[string]any{}

	return c, requests
}

func respondWith(t *testing.T, c *client.Client, requests *[]map[string]any, body string) {
	t.Helper()

	httpmock.RegisterResponder(http.MethodPost, c.GraphqlServerURL, func(req *http.Request) (*http.Response, error) {
		raw, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(raw, &payload))

		*requests = append(*requests, payload)

		return httpmock.NewStringResponse(http.StatusOK, body), nil
	})
}

func configuredResource(t *testing.T, c gatewayResourceCase, cl *client.Client, defaultTags map[string]string) resource.Resource {
	t.Helper()

	res := c.newResource()

	configurable, ok := res.(resource.ResourceWithConfigure)
	require.True(t, ok, "%s must implement Configure", c.name)

	var resp resource.ConfigureResponse

	configurable.Configure(t.Context(), resource.ConfigureRequest{
		ProviderData: &providerdata.ProviderData{Client: cl, DefaultTags: defaultTags},
	}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	return res
}

func schemaOf(t *testing.T, res resource.Resource) schema.Schema {
	t.Helper()

	var resp resource.SchemaResponse

	res.Schema(t.Context(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	return resp.Schema
}

func tagsValue(tags map[string]string) tftypes.Value {
	mapType := tftypes.Map{ElementType: tftypes.String}

	if tags == nil {
		return tftypes.NewValue(mapType, nil)
	}

	elems := make(map[string]tftypes.Value, len(tags))
	for k, v := range tags {
		elems[k] = tftypes.NewValue(tftypes.String, v)
	}

	return tftypes.NewValue(mapType, elems)
}

// objectValue builds a fully-typed object for the schema: attributes not listed
// are null, mirroring how Terraform hands unset attributes to the provider.
func objectValue(t *testing.T, s schema.Schema, attrs map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	objType, ok := s.Type().TerraformType(t.Context()).(tftypes.Object)
	require.True(t, ok)

	values := make(map[string]tftypes.Value, len(objType.AttributeTypes))

	for name, typ := range objType.AttributeTypes {
		if val, found := attrs[name]; found {
			values[name] = val
		} else {
			values[name] = tftypes.NewValue(typ, nil)
		}
	}

	return tftypes.NewValue(objType, values)
}

func nullObject(t *testing.T, s schema.Schema) tftypes.Value {
	t.Helper()

	return tftypes.NewValue(s.Type().TerraformType(t.Context()), nil)
}

func mergeAttrs(base map[string]tftypes.Value, extra map[string]tftypes.Value) map[string]tftypes.Value {
	merged := make(map[string]tftypes.Value, len(base)+len(extra))

	for k, v := range base {
		merged[k] = v
	}

	for k, v := range extra {
		merged[k] = v
	}

	return merged
}

func mapAttribute(t *testing.T, get func(context.Context, path.Path, any) error, name string) types.Map {
	t.Helper()

	var val types.Map

	require.NoError(t, get(t.Context(), path.Root(name), &val))

	return val
}

func stateMap(t *testing.T, state tfsdk.State, name string) types.Map {
	t.Helper()

	return mapAttribute(t, func(ctx context.Context, p path.Path, target any) error {
		diags := state.GetAttribute(ctx, p, target)
		if diags.HasError() {
			return fmt.Errorf("%v", diags) //nolint:err113
		}

		return nil
	}, name)
}

func planMap(t *testing.T, plan tfsdk.Plan, name string) types.Map {
	t.Helper()

	return mapAttribute(t, func(ctx context.Context, p path.Path, target any) error {
		diags := plan.GetAttribute(ctx, p, target)
		if diags.HasError() {
			return fmt.Errorf("%v", diags) //nolint:err113
		}

		return nil
	}, name)
}

func TestGatewayResourcesExposeTagsAll(t *testing.T) {
	for _, c := range gatewayResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			s := schemaOf(t, c.newResource())

			tagsAll, ok := s.Attributes[attr.TagsAll].(schema.MapAttribute)
			require.True(t, ok, "tags_all must be a map attribute")
			assert.True(t, tagsAll.Computed, "tags_all must be computed")
			assert.False(t, tagsAll.Optional, "tags_all must not be user-settable")
			assert.Equal(t, types.StringType, tagsAll.ElementType)
		})
	}
}

// runModifyPlan drives ModifyPlan with a config, plan and (optional) state that
// share the same attribute values apart from the tags supplied by the caller.
func runModifyPlan(t *testing.T, c gatewayResourceCase, defaultTags map[string]string, configTags, stateTags tftypes.Value, hasState bool) tfsdk.Plan {
	t.Helper()

	res := configuredResource(t, c, nil, defaultTags)
	s := schemaOf(t, res)

	config := tfsdk.Config{Schema: s, Raw: objectValue(t, s, mergeAttrs(c.requiredAttrs, map[string]tftypes.Value{attr.Tags: configTags}))}
	plan := tfsdk.Plan{Schema: s, Raw: objectValue(t, s, mergeAttrs(c.requiredAttrs, map[string]tftypes.Value{
		attr.Tags:    configTags,
		attr.TagsAll: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
	}))}

	state := tfsdk.State{Schema: s, Raw: nullObject(t, s)}
	if hasState {
		state.Raw = objectValue(t, s, mergeAttrs(c.requiredAttrs, map[string]tftypes.Value{
			attr.ID:      tftypes.NewValue(tftypes.String, "res-1"),
			attr.Tags:    stateTags,
			attr.TagsAll: stateTags,
		}))
	}

	modifier, ok := res.(resource.ResourceWithModifyPlan)
	require.True(t, ok, "%s must implement ModifyPlan", c.name)

	resp := &resource.ModifyPlanResponse{Plan: plan}

	modifier.ModifyPlan(t.Context(), resource.ModifyPlanRequest{Config: config, Plan: plan, State: state}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	return resp.Plan
}

func TestGatewayResourcesPlanTagsAll(t *testing.T) {
	defaults := map[string]string{"env": "stage", "app": "default_app"}

	for _, c := range gatewayResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Run("config tags merged with provider defaults, config wins", func(t *testing.T) {
				plan := runModifyPlan(t, c, defaults, tagsValue(map[string]string{"owner": "team", "app": "custom_app"}), tagsValue(nil), false)

				assert.Equal(t, stringMap(map[string]string{"env": "stage", "app": "custom_app", "owner": "team"}), planMap(t, plan, attr.TagsAll))
			})

			t.Run("config null falls back to user tags kept in state", func(t *testing.T) {
				plan := runModifyPlan(t, c, defaults, tagsValue(nil), tagsValue(map[string]string{"owner": "team"}), true)

				assert.Equal(t, stringMap(map[string]string{"env": "stage", "app": "default_app", "owner": "team"}), planMap(t, plan, attr.TagsAll))
			})

			t.Run("no defaults and no tags leaves tags_all null", func(t *testing.T) {
				plan := runModifyPlan(t, c, nil, tagsValue(nil), tagsValue(nil), false)

				assert.True(t, planMap(t, plan, attr.TagsAll).IsNull())
			})

			t.Run("defaults only populate tags_all on create", func(t *testing.T) {
				plan := runModifyPlan(t, c, defaults, tagsValue(nil), tagsValue(nil), false)

				assert.Equal(t, stringMap(defaults), planMap(t, plan, attr.TagsAll))
			})

			t.Run("destroy plan is left untouched", func(t *testing.T) {
				res := configuredResource(t, c, nil, defaults)
				s := schemaOf(t, res)

				plan := tfsdk.Plan{Schema: s, Raw: nullObject(t, s)}
				state := tfsdk.State{Schema: s, Raw: objectValue(t, s, mergeAttrs(c.requiredAttrs, map[string]tftypes.Value{
					attr.ID: tftypes.NewValue(tftypes.String, "res-1"),
				}))}

				modifier, ok := res.(resource.ResourceWithModifyPlan)
				require.True(t, ok)

				resp := &resource.ModifyPlanResponse{Plan: plan}
				modifier.ModifyPlan(t.Context(), resource.ModifyPlanRequest{Config: tfsdk.Config{Schema: s, Raw: nullObject(t, s)}, Plan: plan, State: state}, resp)

				require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
				assert.True(t, resp.Plan.Raw.IsNull())
			})
		})
	}
}

func TestGatewayResourcesImportSplitsTags(t *testing.T) {
	// `application` is a default key the user overrode with a different value: the
	// API value can only have come from the user, so it must stay in `tags`.
	defaults := map[string]string{"env": "prod", "application": "default_app"}
	apiTags := map[string]string{"env": "prod", "owner": "team", "application": "custom_app"}

	for _, c := range gatewayResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			cl, requests := mockedClient(t)
			respondWith(t, cl, requests, c.readResponse("res-1", apiTags))

			res := configuredResource(t, c, cl, defaults)
			s := schemaOf(t, res)

			importer, ok := res.(resource.ResourceWithImportState)
			require.True(t, ok)

			resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: s, Raw: nullObject(t, s)}}
			importer.ImportState(t.Context(), resource.ImportStateRequest{ID: "res-1"}, resp)
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			var id types.String

			require.False(t, resp.State.GetAttribute(t.Context(), path.Root(attr.ID), &id).HasError())
			assert.Equal(t, "res-1", id.ValueString())

			assert.Equal(t, stringMap(map[string]string{"owner": "team", "application": "custom_app"}), stateMap(t, resp.State, attr.Tags),
				"tags must exclude provider defaults but keep user overrides of default keys")
			assert.Equal(t, stringMap(apiTags), stateMap(t, resp.State, attr.TagsAll), "tags_all must hold every API tag")
		})
	}
}

func TestGatewayResourcesReadKeepsUserTags(t *testing.T) {
	userTags := map[string]string{"owner": "team"}
	apiTags := map[string]string{"env": "prod", "owner": "team", "added": "outside"}

	for _, c := range gatewayResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			cl, requests := mockedClient(t)
			respondWith(t, cl, requests, c.readResponse("res-1", apiTags))

			res := configuredResource(t, c, cl, map[string]string{"env": "prod"})
			s := schemaOf(t, res)

			state := tfsdk.State{Schema: s, Raw: objectValue(t, s, mergeAttrs(c.requiredAttrs, map[string]tftypes.Value{
				attr.ID:      tftypes.NewValue(tftypes.String, "res-1"),
				attr.Tags:    tagsValue(userTags),
				attr.TagsAll: tagsValue(map[string]string{"env": "prod", "owner": "team"}),
			}))}

			resp := &resource.ReadResponse{State: state}
			res.Read(t.Context(), resource.ReadRequest{State: state}, resp)
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			assert.Equal(t, stringMap(userTags), stateMap(t, resp.State, attr.Tags), "tags must stay as declared by the user")
			assert.Equal(t, stringMap(apiTags), stateMap(t, resp.State, attr.TagsAll), "tags_all must reflect the API")
		})
	}
}

func TestGatewayResourcesCreateSendsTagsAll(t *testing.T) {
	userTags := map[string]string{"owner": "team"}
	allTags := map[string]string{"env": "prod", "owner": "team"}

	for _, c := range gatewayResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			cl, requests := mockedClient(t)
			respondWith(t, cl, requests, c.createResponse("res-1", allTags))

			res := configuredResource(t, c, cl, map[string]string{"env": "prod"})
			s := schemaOf(t, res)

			plan := tfsdk.Plan{Schema: s, Raw: objectValue(t, s, mergeAttrs(c.requiredAttrs, map[string]tftypes.Value{
				attr.Tags:    tagsValue(userTags),
				attr.TagsAll: tagsValue(allTags),
			}))}

			resp := &resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: nullObject(t, s)}}
			res.Create(t.Context(), resource.CreateRequest{Plan: plan}, resp)
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			require.NotEmpty(t, *requests)
			variables, ok := (*requests)[0]["variables"].(map[string]any)
			require.True(t, ok)

			assert.ElementsMatch(t, []any{
				map[string]any{"key": "env", "value": "prod"},
				map[string]any{"key": "owner", "value": "team"},
			}, variables["tags"], "the API must receive provider defaults merged with user tags")

			assert.Equal(t, stringMap(userTags), stateMap(t, resp.State, attr.Tags))
			assert.Equal(t, stringMap(allTags), stateMap(t, resp.State, attr.TagsAll))
		})
	}
}
