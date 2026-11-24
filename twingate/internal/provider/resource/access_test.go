package resource

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	accessAddResponse    = `{"data":{"resourceAccessAdd":{"ok":true,"error":null}}}`
	accessRemoveResponse = `{"data":{"resourceAccessRemove":{"ok":true,"error":null}}}`
)

var errTestReadAccess = errors.New("read access failed")

func testGroupID(id string) string {
	return base64.StdEncoding.EncodeToString([]byte("Group:" + id))
}

func testServiceAccountID(id string) string {
	return base64.StdEncoding.EncodeToString([]byte("ServiceAccount:" + id))
}

func makeTestGroupAccess(ctx context.Context, groupIDs ...string) types.Set {
	if len(groupIDs) == 0 {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes())
	}

	groups := make([]types.Object, 0, len(groupIDs))
	for _, id := range groupIDs {
		groups = append(groups, makeTestAccessGroup(ctx, id, types.StringNull()))
	}

	return makeTestAccessGroupSet(ctx, groups...)
}

func makeTestServiceAccess(ctx context.Context, serviceAccountIDs ...string) types.Set {
	if len(serviceAccountIDs) == 0 {
		return makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes())
	}

	elements := make([]tfattr.Value, 0, len(serviceAccountIDs))
	for _, id := range serviceAccountIDs {
		elements = append(elements, types.ObjectValueMust(accessServiceAccountAttributeTypes(), map[string]tfattr.Value{
			attr.ServiceAccountID: types.StringValue(id),
		}))
	}

	return types.SetValueMust(types.ObjectNull(accessServiceAccountAttributeTypes()).Type(ctx), elements)
}

func makeTestAccessModel(ctx context.Context, isAuthoritative types.Bool, groupIDs, serviceAccountIDs []string) accessModel {
	return accessModel{
		IsAuthoritative: isAuthoritative,
		AccessPolicy:    makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
		GroupAccess:     makeTestGroupAccess(ctx, groupIDs...),
		ServiceAccess:   makeTestServiceAccess(ctx, serviceAccountIDs...),
	}
}

func groupIDsOf(groups []model.AccessGroup) []string {
	ids := make([]string, 0, len(groups))
	for _, group := range groups {
		ids = append(ids, group.GroupID)
	}

	return ids
}

func testGroups(ids ...string) []model.AccessGroup {
	groups := make([]model.AccessGroup, 0, len(ids))
	for _, id := range ids {
		groups = append(groups, model.AccessGroup{GroupID: id})
	}

	return groups
}

func TestIsAccessAuthoritative(t *testing.T) {
	// State written before `is_authoritative` existed holds null: it must keep the default.
	assert.True(t, isAccessAuthoritative(types.BoolNull()))
	assert.True(t, isAccessAuthoritative(types.BoolUnknown()))
	assert.True(t, isAccessAuthoritative(types.BoolValue(true)))
	assert.False(t, isAccessAuthoritative(types.BoolValue(false)))
}

func TestParseAccess(t *testing.T) {
	ctx := t.Context()
	group, serviceAccount := testGroupID("1"), testServiceAccountID("1")

	cases := []struct {
		name        string
		plan        accessModel
		expected    resourceAccess
		expectedErr error
	}{
		{
			name: "no access",
			plan: makeTestAccessModel(ctx, types.BoolValue(true), nil, nil),
		},
		{
			name:     "groups and service accounts",
			plan:     makeTestAccessModel(ctx, types.BoolValue(true), []string{group}, []string{serviceAccount}),
			expected: resourceAccess{Groups: testGroups(group), ServiceAccounts: []string{serviceAccount}},
		},
		{
			name:        "empty group ID",
			plan:        makeTestAccessModel(ctx, types.BoolValue(true), []string{""}, nil),
			expectedErr: ErrInvalidAttributeCombination,
		},
		{
			name:        "group ID is not a global ID",
			plan:        makeTestAccessModel(ctx, types.BoolValue(true), []string{"group-1"}, nil),
			expectedErr: ErrWrongGlobalID,
		},
		{
			name:        "service account ID is not a global ID",
			plan:        makeTestAccessModel(ctx, types.BoolValue(true), nil, []string{"service-account-1"}),
			expectedErr: ErrWrongGlobalID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			access, accessPolicy, err := parseAccess(c.plan)

			if c.expectedErr != nil {
				require.ErrorIs(t, err, c.expectedErr)

				return
			}

			require.NoError(t, err)
			assert.Nil(t, accessPolicy)
			assert.Equal(t, c.expected, access)
		})
	}
}

func TestParseAccessInvalidAccessPolicy(t *testing.T) {
	ctx := t.Context()

	plan := makeTestAccessModel(ctx, types.BoolValue(true), nil, nil)
	policy := types.ObjectValueMust(accessPolicyAttributeTypes(), map[string]tfattr.Value{
		attr.Mode:         types.StringValue(model.AccessPolicyModeAutoLock),
		attr.Duration:     types.StringValue("not-a-duration"),
		attr.ApprovalMode: types.StringNull(),
	})

	var diags diag.Diagnostics

	plan.AccessPolicy, diags = makeObjectsSet(ctx, policy)
	require.False(t, diags.HasError(), diags)

	_, _, err := parseAccess(plan)
	require.ErrorContains(t, err, "failed to parse access_policy")
}

func TestGetChangedAccess(t *testing.T) {
	ctx := t.Context()

	cases := []struct {
		name            string
		isAuthoritative bool
		stateGroups     []string
		stateServices   []string
		planGroups      []string
		planServices    []string
		remote          resourceAccess
		expectedDelete  []string
		expectedAdd     resourceAccess
	}{
		{
			name:            "authoritative - removes everything remote that is not planned",
			isAuthoritative: true,
			stateGroups:     []string{"g1"},
			stateServices:   []string{"s1"},
			planGroups:      []string{"g1", "g2"},
			planServices:    []string{"s2"},
			remote:          resourceAccess{Groups: testGroups("g1", "g-external"), ServiceAccounts: []string{"s1", "s-external"}},
			expectedDelete:  []string{"g-external", "s1", "s-external"},
			expectedAdd:     resourceAccess{Groups: testGroups("g2"), ServiceAccounts: []string{"s2"}},
		},
		{
			name:            "non-authoritative - removes only what Terraform assigned",
			isAuthoritative: false,
			stateGroups:     []string{"g1"},
			stateServices:   []string{"s1"},
			planGroups:      []string{"g2"},
			planServices:    []string{"s2"},
			remote:          resourceAccess{Groups: testGroups("g1", "g-external"), ServiceAccounts: []string{"s1", "s-external"}},
			expectedDelete:  []string{"g1", "s1"},
			expectedAdd:     resourceAccess{Groups: testGroups("g2"), ServiceAccounts: []string{"s2"}},
		},
		{
			name:            "non-authoritative - unchanged block is left alone",
			isAuthoritative: false,
			stateGroups:     []string{"g1"},
			stateServices:   []string{"s1"},
			planGroups:      []string{"g1"},
			planServices:    nil,
			remote:          resourceAccess{Groups: testGroups("g1", "g-external"), ServiceAccounts: []string{"s1", "s-external"}},
			expectedDelete:  []string{"s1"},
			expectedAdd:     resourceAccess{Groups: []model.AccessGroup{}},
		},
		{
			name:            "already assigned remotely - nothing is added twice",
			isAuthoritative: true,
			planGroups:      []string{"g1"},
			planServices:    []string{"s1"},
			remote:          resourceAccess{Groups: testGroups("g1"), ServiceAccounts: []string{"s1"}},
			expectedAdd:     resourceAccess{Groups: []model.AccessGroup{}, ServiceAccounts: []string{}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan := makeTestAccessModel(ctx, types.BoolValue(c.isAuthoritative), c.planGroups, c.planServices)
			state := makeTestAccessModel(ctx, types.BoolValue(c.isAuthoritative), c.stateGroups, c.stateServices)
			desired := resourceAccess{Groups: testGroups(c.planGroups...), ServiceAccounts: c.planServices}

			idsToDelete, toAdd, err := getChangedAccess(c.isAuthoritative, plan, state, desired, c.remote)

			require.NoError(t, err)
			assert.ElementsMatch(t, c.expectedDelete, idsToDelete)
			assert.ElementsMatch(t, c.expectedAdd.Groups, toAdd.Groups)
			assert.ElementsMatch(t, c.expectedAdd.ServiceAccounts, toAdd.ServiceAccounts)
		})
	}
}

func TestFilterNonAuthoritativeAccess(t *testing.T) {
	ctx := t.Context()

	reference := makeTestAccessModel(ctx, types.BoolValue(false), []string{"g1", "g-missing"}, []string{"s1", "s-missing"})
	remote := resourceAccess{Groups: testGroups("g1", "g-external"), ServiceAccounts: []string{"s1", "s-external"}}

	access, err := filterNonAuthoritativeAccess(reference, remote)

	require.NoError(t, err)
	assert.Equal(t, []string{"g1"}, groupIDsOf(access.Groups))
	assert.Equal(t, []string{"s1"}, access.ServiceAccounts)
}

func TestUpdateResourceAccess(t *testing.T) {
	ctx := t.Context()

	t.Run("unchanged access - API is not called", func(t *testing.T) {
		apiClient, requests := newSequencedMockedClient(t)
		access := makeTestAccessModel(ctx, types.BoolValue(true), []string{"g1"}, []string{"s1"})

		err := updateResourceAccess(ctx, apiClient, testResourceID, true, access, access, resourceAccess{}, func(context.Context) (resourceAccess, error) {
			t.Error("remote access must not be read")

			return resourceAccess{}, nil
		})

		require.NoError(t, err)
		assert.Empty(t, *requests)
	})

	t.Run("read failure is reported", func(t *testing.T) {
		apiClient, requests := newSequencedMockedClient(t)
		plan := makeTestAccessModel(ctx, types.BoolValue(true), nil, []string{"s1"})
		state := makeTestAccessModel(ctx, types.BoolValue(true), nil, nil)

		err := updateResourceAccess(ctx, apiClient, testResourceID, true, plan, state, resourceAccess{ServiceAccounts: []string{"s1"}}, func(context.Context) (resourceAccess, error) {
			return resourceAccess{}, errTestReadAccess
		})

		require.ErrorIs(t, err, errTestReadAccess)
		assert.Empty(t, *requests)
	})

	t.Run("changed access - removes then adds", func(t *testing.T) {
		apiClient, requests := newSequencedMockedClient(t, accessRemoveResponse, accessAddResponse)
		plan := makeTestAccessModel(ctx, types.BoolValue(true), []string{"g1"}, []string{"s2"})
		state := makeTestAccessModel(ctx, types.BoolValue(true), []string{"g1"}, []string{"s1"})
		desired := resourceAccess{Groups: testGroups("g1"), ServiceAccounts: []string{"s2"}}

		err := updateResourceAccess(ctx, apiClient, testResourceID, true, plan, state, desired, func(context.Context) (resourceAccess, error) {
			return resourceAccess{Groups: testGroups("g1"), ServiceAccounts: []string{"s1"}}, nil
		})

		require.NoError(t, err)
		require.Len(t, *requests, 2)
		assert.Equal(t, []any{"s1"}, requestVariable(t, (*requests)[0], "principalIds"))
		assert.Equal(t, []string{"s2"}, accessPrincipalIDs(t, (*requests)[1]))
	})

	t.Run("API failure is reported", func(t *testing.T) {
		apiClient, _ := newSequencedMockedClient(t, graphqlErrorResponse)
		plan := makeTestAccessModel(ctx, types.BoolValue(true), nil, []string{"s2"})
		state := makeTestAccessModel(ctx, types.BoolValue(true), nil, nil)

		err := updateResourceAccess(ctx, apiClient, testResourceID, true, plan, state, resourceAccess{ServiceAccounts: []string{"s2"}}, func(context.Context) (resourceAccess, error) {
			return resourceAccess{}, nil
		})

		require.ErrorContains(t, err, "failed to update resource access")
	})
}

func TestRefreshAccess(t *testing.T) {
	ctx := t.Context()
	remote := resourceAccess{Groups: testGroups("g1", "g-external"), ServiceAccounts: []string{"s1", "s-external"}}

	t.Run("null is_authoritative - defaults to authoritative", func(t *testing.T) {
		var diags diag.Diagnostics

		reference := makeTestAccessModel(ctx, types.BoolNull(), []string{"g1"}, nil)
		access := refreshAccess(ctx, reference, nil, remote, &diags)

		require.False(t, diags.HasError(), diags)
		assert.Equal(t, types.BoolValue(true), access.IsAuthoritative)
		assert.ElementsMatch(t, []string{"s1", "s-external"}, getServiceAccountAccessAttribute(access.ServiceAccess))

		groups, err := getGroupAccessAttribute(access.GroupAccess)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"g1", "g-external"}, groupIDsOf(groups))
	})

	t.Run("non-authoritative - external assignments are ignored", func(t *testing.T) {
		var diags diag.Diagnostics

		reference := makeTestAccessModel(ctx, types.BoolValue(false), []string{"g1"}, []string{"s1"})
		access := refreshAccess(ctx, reference, nil, remote, &diags)

		require.False(t, diags.HasError(), diags)
		assert.Equal(t, types.BoolValue(false), access.IsAuthoritative)
		assert.Equal(t, reference.GroupAccess, access.GroupAccess)
		assert.Equal(t, reference.ServiceAccess, access.ServiceAccess)
	})
}

// accessPrincipalIDs extracts the principal IDs, in order, from a resourceAccessAdd request.
func accessPrincipalIDs(t *testing.T, requestBody string) []string {
	t.Helper()

	inputs, ok := requestVariable(t, requestBody, "access").([]any)
	require.True(t, ok, requestBody)

	ids := make([]string, 0, len(inputs))

	for _, input := range inputs {
		entry, ok := input.(map[string]any)
		require.True(t, ok, requestBody)

		id, ok := entry["principalId"].(string)
		require.True(t, ok, requestBody)

		ids = append(ids, id)
	}

	return ids
}

// pamModel is implemented by the models of all PAM resources.
type pamModel interface {
	access() accessModel
	setAccess(access accessModel)
}

type pamResourceCase struct {
	name        string
	newResource func(apiClient *client.Client) resource.ResourceWithImportState
	// newModel returns a valid plan without any access.
	newModel   func(t *testing.T) pamModel
	emptyModel func() pamModel
	// entityResponse is the answer of the create or update mutation.
	entityResponse func(t *testing.T, create bool) string
}

func pamResourceCases() []pamResourceCase {
	mutation := func(prefix string, create bool) string {
		if create {
			return prefix + "Create"
		}

		return prefix + "Update"
	}

	return []pamResourceCase{
		{
			name: TwingateKubernetesResource,
			newResource: func(apiClient *client.Client) resource.ResourceWithImportState {
				return &kubernetesResource{client: apiClient}
			},
			newModel: func(t *testing.T) pamModel {
				t.Helper()

				plan := testKubernetesModel(t, types.MapNull(types.StringType))

				return &plan
			},
			emptyModel: func() pamModel { return &kubernetesResourceModel{} },
			entityResponse: func(t *testing.T, create bool) string {
				t.Helper()

				return kubernetesEntityResponse(t, mutation("kubernetesResource", create), nil)
			},
		},
		{
			name: TwingateSSHResource,
			newResource: func(apiClient *client.Client) resource.ResourceWithImportState {
				return &sshResource{client: apiClient}
			},
			newModel: func(t *testing.T) pamModel {
				t.Helper()

				plan := testSSHModel(t, types.MapNull(types.StringType))

				return &plan
			},
			emptyModel: func() pamModel { return &sshResourceModel{} },
			entityResponse: func(t *testing.T, create bool) string {
				t.Helper()

				return sshEntityResponse(t, mutation("sshResource", create), nil, defaultSSHPort, defaultSSHPort)
			},
		},
		{
			name: TwingateWebAppResource,
			newResource: func(apiClient *client.Client) resource.ResourceWithImportState {
				return &webAppResource{client: apiClient}
			},
			newModel: func(t *testing.T) pamModel {
				t.Helper()

				plan := testWebAppModel(t, types.MapNull(types.StringType), types.MapNull(types.StringType))

				return &plan
			},
			emptyModel: func() pamModel { return &webAppResourceModel{} },
			entityResponse: func(t *testing.T, create bool) string {
				t.Helper()

				return webAppEntityResponse(t, mutation("webAppResource", create), nil, nil)
			},
		},
	}
}

// readResponse turns the entity of a mutation response into the response of
// the read query, with the given principals assigned and the API default access policy.
func (c pamResourceCase) readResponse(t *testing.T, groupIDs, serviceAccountIDs []string) string {
	t.Helper()

	var mutation struct {
		Data map[string]struct {
			Entity map[string]any `json:"entity"`
		} `json:"data"`
	}

	require.NoError(t, json.Unmarshal([]byte(c.entityResponse(t, true)), &mutation))
	require.Len(t, mutation.Data, 1)

	edges := make([]any, 0, len(groupIDs)+len(serviceAccountIDs))

	for _, id := range groupIDs {
		edges = append(edges, map[string]any{"node": map[string]any{"__typename": "Group", "id": id}})
	}

	for _, id := range serviceAccountIDs {
		edges = append(edges, map[string]any{"node": map[string]any{"__typename": "ServiceAccount", "id": id}})
	}

	for _, payload := range mutation.Data {
		payload.Entity["approvalMode"] = model.ApprovalModeManual
		payload.Entity["accessPolicy"] = map[string]any{"mode": model.AccessPolicyModeManual, "durationSeconds": nil}
		payload.Entity["access"] = map[string]any{
			"pageInfo": map[string]any{"endCursor": "", "hasNextPage": false},
			"edges":    edges,
		}

		raw, err := json.Marshal(map[string]any{"data": map[string]any{"resource": payload.Entity}})
		require.NoError(t, err)

		return string(raw)
	}

	return ""
}

func (c pamResourceCase) model(t *testing.T, id types.String, isAuthoritative types.Bool, groupIDs, serviceAccountIDs []string) pamModel {
	t.Helper()

	plan := c.newModel(t)
	plan.setAccess(makeTestAccessModel(t.Context(), isAuthoritative, groupIDs, serviceAccountIDs))

	// every PAM model starts with the ID field
	switch typed := plan.(type) {
	case *kubernetesResourceModel:
		typed.ID = id
	case *sshResourceModel:
		typed.ID = id
	case *webAppResourceModel:
		typed.ID = id
	}

	return plan
}

func assertAccessState(t *testing.T, access accessModel, isAuthoritative bool, groupIDs, serviceAccountIDs []string) {
	t.Helper()

	assert.Equal(t, types.BoolValue(isAuthoritative), access.IsAuthoritative)

	groups, err := getGroupAccessAttribute(access.GroupAccess)
	require.NoError(t, err)
	assert.ElementsMatch(t, groupIDs, groupIDsOf(groups))
	assert.ElementsMatch(t, serviceAccountIDs, getServiceAccountAccessAttribute(access.ServiceAccess))
}

// Service accounts and groups are assigned right after the resource is created,
// in one resourceAccessAdd call, and land in state.
func TestPAMResourceCreateAssignsAccess(t *testing.T) {
	group, serviceAccount := testGroupID("1"), testServiceAccountID("1")

	for _, c := range pamResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			apiClient, requests := newSequencedMockedClient(t, c.entityResponse(t, true), accessAddResponse)
			res := c.newResource(apiClient)

			plan := c.model(t, types.StringUnknown(), types.BoolValue(true), []string{group}, []string{serviceAccount})

			resp := &resource.CreateResponse{State: nullStateOf(t, res)}
			res.Create(t.Context(), resource.CreateRequest{Plan: planOf(t, res, plan)}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Len(t, *requests, 2)
			assert.Equal(t, []string{serviceAccount, group}, accessPrincipalIDs(t, (*requests)[1]))

			state := c.emptyModel()
			readStateInto(t, resp.State, state)
			assertAccessState(t, state.access(), true, []string{group}, []string{serviceAccount})
		})
	}
}

// Access IDs are validated before anything is sent to the API.
func TestPAMResourceRejectsInvalidAccessIDs(t *testing.T) {
	for _, c := range pamResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			apiClient, requests := newSequencedMockedClient(t)
			res := c.newResource(apiClient)

			plan := c.model(t, types.StringUnknown(), types.BoolValue(true), nil, []string{"not-a-global-id"})

			resp := &resource.CreateResponse{State: nullStateOf(t, res)}
			res.Create(t.Context(), resource.CreateRequest{Plan: planOf(t, res, plan)}, resp)

			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), ErrWrongGlobalID.Error())
			assert.Empty(t, *requests)
		})
	}
}

func TestPAMResourceUpdateAccess(t *testing.T) {
	group, externalGroup := testGroupID("1"), testGroupID("external")
	oldService, newService, externalService := testServiceAccountID("old"), testServiceAccountID("new"), testServiceAccountID("external")

	cases := []struct {
		name            string
		isAuthoritative bool
		expectedRemoved []any
	}{
		{
			name:            "authoritative - external assignments are removed",
			isAuthoritative: true,
			expectedRemoved: []any{oldService, externalService, externalGroup},
		},
		{
			name:            "non-authoritative - external assignments are kept",
			isAuthoritative: false,
			expectedRemoved: []any{oldService},
		},
	}

	for _, c := range pamResourceCases() {
		for _, mode := range cases {
			t.Run(c.name+": "+mode.name, func(t *testing.T) {
				apiClient, requests := newSequencedMockedClient(t,
					c.readResponse(t, []string{group, externalGroup}, []string{oldService, externalService}),
					accessRemoveResponse,
					accessAddResponse,
					c.entityResponse(t, false),
				)
				res := c.newResource(apiClient)

				isAuthoritative := types.BoolValue(mode.isAuthoritative)
				stateModel := c.model(t, types.StringValue(testResourceID), isAuthoritative, []string{group}, []string{oldService})
				planModel := c.model(t, types.StringValue(testResourceID), isAuthoritative, []string{group}, []string{newService})

				resp := &resource.UpdateResponse{State: stateOf(t, res, stateModel)}
				res.Update(t.Context(), resource.UpdateRequest{
					Plan:  planOf(t, res, planModel),
					State: stateOf(t, res, stateModel),
				}, resp)

				require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
				require.Len(t, *requests, 4)
				assert.ElementsMatch(t, mode.expectedRemoved, requestVariable(t, (*requests)[1], "principalIds"))
				assert.Equal(t, []string{newService}, accessPrincipalIDs(t, (*requests)[2]))

				state := c.emptyModel()
				readStateInto(t, resp.State, state)
				assertAccessState(t, state.access(), mode.isAuthoritative, []string{group}, []string{newService})
			})
		}
	}
}

// An update that leaves the access blocks alone must not touch the remote access.
func TestPAMResourceUpdateKeepsUnchangedAccess(t *testing.T) {
	group, serviceAccount := testGroupID("1"), testServiceAccountID("1")

	for _, c := range pamResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			apiClient, requests := newSequencedMockedClient(t, c.entityResponse(t, false))
			res := c.newResource(apiClient)

			// State written before `is_authoritative` existed holds null.
			stateModel := c.model(t, types.StringValue(testResourceID), types.BoolNull(), []string{group}, []string{serviceAccount})
			planModel := c.model(t, types.StringValue(testResourceID), types.BoolValue(true), []string{group}, []string{serviceAccount})

			resp := &resource.UpdateResponse{State: stateOf(t, res, stateModel)}
			res.Update(t.Context(), resource.UpdateRequest{
				Plan:  planOf(t, res, planModel),
				State: stateOf(t, res, stateModel),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Len(t, *requests, 1)

			state := c.emptyModel()
			readStateInto(t, resp.State, state)
			assertAccessState(t, state.access(), true, []string{group}, []string{serviceAccount})
		})
	}
}

func TestPAMResourceReadAccess(t *testing.T) {
	group, externalGroup := testGroupID("1"), testGroupID("external")
	serviceAccount, externalService := testServiceAccountID("1"), testServiceAccountID("external")

	cases := []struct {
		name             string
		isAuthoritative  types.Bool
		expectedFlag     bool
		expectedGroups   []string
		expectedServices []string
	}{
		{
			name:             "state without is_authoritative - upgraded to authoritative",
			isAuthoritative:  types.BoolNull(),
			expectedFlag:     true,
			expectedGroups:   []string{group, externalGroup},
			expectedServices: []string{serviceAccount, externalService},
		},
		{
			name:             "authoritative - external assignments show up as drift",
			isAuthoritative:  types.BoolValue(true),
			expectedFlag:     true,
			expectedGroups:   []string{group, externalGroup},
			expectedServices: []string{serviceAccount, externalService},
		},
		{
			name:             "non-authoritative - external assignments are ignored",
			isAuthoritative:  types.BoolValue(false),
			expectedFlag:     false,
			expectedGroups:   []string{group},
			expectedServices: []string{serviceAccount},
		},
	}

	for _, c := range pamResourceCases() {
		for _, mode := range cases {
			t.Run(c.name+": "+mode.name, func(t *testing.T) {
				apiClient, _ := newMockedClient(t, c.readResponse(t, []string{group, externalGroup}, []string{serviceAccount, externalService}))
				res := c.newResource(apiClient)

				stateModel := c.model(t, types.StringValue(testResourceID), mode.isAuthoritative, []string{group}, []string{serviceAccount})

				resp := &resource.ReadResponse{State: stateOf(t, res, stateModel)}
				res.Read(t.Context(), resource.ReadRequest{State: stateOf(t, res, stateModel)}, resp)

				require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

				state := c.emptyModel()
				readStateInto(t, resp.State, state)
				assertAccessState(t, state.access(), mode.expectedFlag, mode.expectedGroups, mode.expectedServices)
				// access_policy is not configured, so the API default stays out of state
				assert.True(t, state.access().AccessPolicy.IsNull())
			})
		}
	}
}

// Import brings in everything assigned remotely and marks the resource authoritative.
func TestPAMResourceImportAccess(t *testing.T) {
	group, serviceAccount := testGroupID("1"), testServiceAccountID("1")

	for _, c := range pamResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			apiClient, _ := newMockedClient(t, c.readResponse(t, []string{group}, []string{serviceAccount}))
			res := c.newResource(apiClient)

			resp := &resource.ImportStateResponse{State: nullStateOf(t, res)}
			res.ImportState(t.Context(), resource.ImportStateRequest{ID: testResourceID}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			var access accessModel

			for attribute, target := range map[string]any{
				attr.IsAuthoritative: &access.IsAuthoritative,
				attr.AccessPolicy:    &access.AccessPolicy,
				attr.AccessGroup:     &access.GroupAccess,
				attr.AccessService:   &access.ServiceAccess,
			} {
				diags := resp.State.GetAttribute(t.Context(), path.Root(attribute), target)
				require.False(t, diags.HasError(), fmt.Sprintf("%s: %v", attribute, diags))
			}

			assertAccessState(t, access, true, []string{group}, []string{serviceAccount})
			assert.Len(t, access.AccessPolicy.Elements(), 1)
		})
	}
}
