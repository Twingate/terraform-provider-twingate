package acctests

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/datasource"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var (
	ErrResourceIDNotSet            = errors.New("id not set")
	ErrResourceNotFound            = errors.New("resource not found")
	ErrResourceStillPresent        = errors.New("resource still present")
	ErrResourceFoundInState        = errors.New("this resource should not be here")
	ErrUnknownResourceType         = errors.New("unknown resource type")
	ErrClientNotInitialized        = errors.New("meta client not initialized")
	ErrSecurityPoliciesNotFound    = errors.New("security policies not found")
	ErrInvalidPath                 = errors.New("invalid path: the path value cannot be asserted as string")
	ErrNotNullSecurityPolicy       = errors.New("expected null security policy in GroupAccess, got non null")
	ErrNotNullUsageBased           = errors.New("expected null usage based duration in GroupAccess, got non null")
	ErrNullSecurityPolicy          = errors.New("expected non null security policy in GroupAccess, got null")
	ErrNullUsageBased              = errors.New("expected non null usage based duration in GroupAccess, got null")
	ErrEmptyGroupAccess            = errors.New("expected at least one group in GroupAccess")
	ErrNotNullUsageBasedOnResource = errors.New("expected null usage based duration on Resource, got non null")
	ErrEmptyTagsList               = errors.New("expected non-empty list of tags")
)

func ErrServiceAccountsLenMismatch(expected, actual int) error {
	return fmt.Errorf("expected %d service accounts, actual - %d", expected, actual) //nolint
}

func ErrDNSProfileAllowedDomainsLenMismatch(expected, actual int) error {
	return fmt.Errorf("expected %d allowed domains, actual - %d", expected, actual) //nolint
}

func ErrGroupsLenMismatch(expected, actual int) error {
	return fmt.Errorf("expected %d groups, actual - %d", expected, actual) //nolint
}

func ErrUsersLenMismatch(expected, actual int) error {
	return fmt.Errorf("expected %d users, actual - %d", expected, actual) //nolint
}

var providerClient = func() *client.Client { //nolint
	client, err := test.TwingateClient()
	if err != nil {
		log.Fatal("failed to init client:", err)
	}

	return client
}()

var ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){ //nolint
	"twingate": providerserver.NewProtocol6WithError(twingate.New(client.DefaultAgent, "test")()),
}

const WaitDuration = 500 * time.Millisecond

func WaitTestFunc() sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		// Sleep 500 ms
		time.Sleep(WaitDuration)

		return nil
	}
}

func ComposeTestCheckFunc(checkFuncs ...sdk.TestCheckFunc) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		if err := WaitTestFunc()(state); err != nil {
			return fmt.Errorf("WaitTestFunc error: %w", err)
		}

		for i, f := range checkFuncs {
			if err := f(state); err != nil {
				return fmt.Errorf("check %d/%d error: %w", i+1, len(checkFuncs), err)
			}
		}

		return nil
	}
}

func PreCheck(t *testing.T) {
	t.Helper()

	var requiredEnvironmentVariables = []string{
		twingate.EnvAPIToken,
		twingate.EnvNetwork,
		twingate.EnvURL,
	}

	t.Run("Test Twingate Resource : AccPreCheck", func(t *testing.T) {
		for _, requiredEnvironmentVariable := range requiredEnvironmentVariables {
			if value := os.Getenv(requiredEnvironmentVariable); value == "" {
				t.Fatalf("%s must be set before running acceptance tests.", requiredEnvironmentVariable)
			}
		}
	})
}

func CheckTwingateResourceDoesNotExists(resourceName string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return nil
		}

		return fmt.Errorf("%w: %s", ErrResourceFoundInState, resourceName)
	}
}

func CheckTwingateServiceAccountDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateServiceAccount {
			continue
		}

		serviceAccountID := rs.Primary.ID

		serviceAccount, _ := providerClient.ReadShallowServiceAccount(context.Background(), serviceAccountID)
		if serviceAccount != nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, serviceAccountID)
		}
	}

	return nil
}

func CheckTwingateResourceExists(resourceName string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%w: %s", ErrResourceNotFound, resourceName)
		}

		if resource.Primary.ID == "" {
			return ErrResourceIDNotSet
		}

		return nil
	}
}

func GetTwingateResourceID(resourceName string, resourceID **string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%w: %s", ErrResourceNotFound, resourceName)
		}

		if resource.Primary.ID == "" {
			return ErrResourceIDNotSet
		}

		id := resource.Primary.ID
		*resourceID = &id

		return nil
	}
}

func DatasourceName(resource, name string) string {
	return fmt.Sprintf("data.%s.%s", resource, name)
}

func ResourceName(resource, name string) string {
	return fmt.Sprintf("%s.%s", resource, name)
}

func TerraformResource(name string) string {
	return ResourceName(resource.TwingateResource, name)
}

func TerraformRemoteNetwork(name string) string {
	return ResourceName(resource.TwingateRemoteNetwork, name)
}

func TerraformGroup(name string) string {
	return ResourceName(resource.TwingateGroup, name)
}

func TerraformConnector(name string) string {
	return ResourceName(resource.TwingateConnector, name)
}

func TerraformConnectorTokens(name string) string {
	return ResourceName(resource.TwingateConnectorTokens, name)
}

func TerraformServiceAccount(name string) string {
	return ResourceName(resource.TwingateServiceAccount, name)
}

func TerraformServiceKey(name string) string {
	return ResourceName(resource.TwingateServiceAccountKey, name)
}

func TerraformUser(name string) string {
	return ResourceName(resource.TwingateUser, name)
}

func TerraformDNSFilteringProfile(name string) string {
	return ResourceName(resource.TwingateDNSFilteringProfile, name)
}

func TerraformDatasourceUsers(name string) string {
	return DatasourceName(datasource.TwingateUsers, name)
}

func DeleteTwingateResource(resourceName, resourceType string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%w: %s ", ErrResourceNotFound, resourceName)
		}

		resourceID := resourceState.Primary.ID
		if resourceID == "" {
			return ErrResourceIDNotSet
		}

		err := deleteResource(resourceType, resourceID)
		if err != nil {
			return fmt.Errorf("%s with ID %s still active: %w", resourceType, resourceID, err)
		}

		return nil
	}
}

func deleteResource(resourceType, resourceID string) error {
	var err error

	switch resourceType {
	case resource.TwingateResource:
		err = providerClient.DeleteResource(context.Background(), resourceID)
	case resource.TwingateRemoteNetwork:
		err = providerClient.DeleteRemoteNetwork(context.Background(), resourceID)
	case resource.TwingateGroup:
		err = providerClient.DeleteGroup(context.Background(), resourceID)
	case resource.TwingateConnector:
		err = providerClient.DeleteConnector(context.Background(), resourceID)
	case resource.TwingateServiceAccount:
		err = providerClient.DeleteServiceAccount(context.Background(), resourceID)
	case resource.TwingateServiceAccountKey:
		err = providerClient.DeleteServiceKey(context.Background(), resourceID)
	case resource.TwingateUser:
		err = providerClient.DeleteUser(context.Background(), resourceID)
	default:
		err = fmt.Errorf("%s %w", resourceType, ErrUnknownResourceType)
	}

	return err
}

func CheckTwingateResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateResource {
			continue
		}

		resourceID := rs.Primary.ID

		resource, _ := providerClient.ReadResource(context.Background(), resourceID)
		if resource != nil {
			return fmt.Errorf("%w: with ID %s", ErrResourceStillPresent, resourceID)
		}
	}

	return nil
}

func DeactivateTwingateResource(resourceName string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%w: %s", ErrResourceNotFound, resourceName)
		}

		resourceID := resourceState.Primary.ID

		if resourceID == "" {
			return ErrResourceIDNotSet
		}

		err := providerClient.UpdateResourceActiveState(context.Background(), &model.Resource{
			ID:       resourceID,
			IsActive: false,
		})
		if err != nil {
			return fmt.Errorf("resource with ID %s still active: %w", resourceID, err)
		}

		return nil
	}
}

func CheckTwingateResource(resourceName string, check func(res *model.Resource) error) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%w: %s", ErrResourceNotFound, resourceName)
		}

		if resourceState.Primary.ID == "" {
			return ErrResourceIDNotSet
		}

		res, err := providerClient.ReadResource(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read resource: %w", err)
		}

		return check(res)
	}
}

func CheckTwingateResourceSecurityPolicyOnGroupAccess(resourceName string, expectedSecurityPolicy string) sdk.TestCheckFunc {
	return CheckTwingateResource(resourceName, func(res *model.Resource) error {
		if len(res.GroupsAccess) == 0 {
			return ErrEmptyGroupAccess
		}

		if res.GroupsAccess[0].SecurityPolicyID == nil {
			return ErrNullSecurityPolicy
		}

		if *res.GroupsAccess[0].SecurityPolicyID != expectedSecurityPolicy {
			return fmt.Errorf("expected security policy %v, got %v", expectedSecurityPolicy, *res.GroupsAccess[0].SecurityPolicyID) //nolint:err113
		}

		return nil
	})
}

func CheckTwingateResourceUsageBasedOnGroupAccess(resourceName string, expectedUsageBased int64) sdk.TestCheckFunc {
	return CheckTwingateResource(resourceName, func(res *model.Resource) error {
		if len(res.GroupsAccess) == 0 {
			return ErrEmptyGroupAccess
		}

		if res.GroupsAccess[0].UsageBasedDuration == nil {
			return ErrNullUsageBased
		}

		if *res.GroupsAccess[0].UsageBasedDuration != expectedUsageBased {
			return fmt.Errorf("expected usage based duration %v, got %v", expectedUsageBased, *res.GroupsAccess[0].UsageBasedDuration) //nolint:err113
		}

		return nil
	})
}

func CheckTwingateResourceUsageBasedDuration(resourceName string, expectedUsageBased int64) sdk.TestCheckFunc {
	return CheckTwingateResource(resourceName, func(res *model.Resource) error {
		if res.UsageBasedAutolockDurationDays == nil {
			return fmt.Errorf("expected usage based duration %v, got <nil>", expectedUsageBased) //nolint:err113
		}

		if *res.UsageBasedAutolockDurationDays != expectedUsageBased {
			return fmt.Errorf("expected usage based duration %v, got %v", expectedUsageBased, *res.UsageBasedAutolockDurationDays) //nolint:err113
		}

		return nil
	})
}

func CheckTwingateResourceTags(resourceName, tag, expectedValue string) sdk.TestCheckFunc {
	return CheckTwingateResource(resourceName, func(res *model.Resource) error {
		if len(res.Tags) == 0 {
			return ErrEmptyTagsList
		}

		if res.Tags[tag] != expectedValue {
			return fmt.Errorf("expected tag value %v, got %v", expectedValue, res.Tags[tag]) //nolint:err113
		}

		return nil
	})
}

func CheckTwingateResourceSecurityPolicyIsNullOnGroupAccess(resourceName string) sdk.TestCheckFunc {
	return CheckTwingateResource(resourceName, func(res *model.Resource) error {
		if len(res.GroupsAccess) == 0 {
			return ErrEmptyGroupAccess
		}

		if res.GroupsAccess[0].SecurityPolicyID != nil {
			return ErrNotNullSecurityPolicy
		}

		return nil
	})
}

func CheckTwingateResourceUsageBasedIsNullOnGroupAccess(resourceName string) sdk.TestCheckFunc {
	return CheckTwingateResource(resourceName, func(res *model.Resource) error {
		if len(res.GroupsAccess) == 0 {
			return ErrEmptyGroupAccess
		}

		if res.GroupsAccess[0].UsageBasedDuration != nil {
			return ErrNotNullUsageBased
		}

		return nil
	})
}

func CheckTwingateResourceUsageBasedIsNullOnResource(resourceName string) sdk.TestCheckFunc {
	return CheckTwingateResource(resourceName, func(res *model.Resource) error {
		if res.UsageBasedAutolockDurationDays != nil {
			return ErrNotNullUsageBasedOnResource
		}

		return nil
	})
}

func CheckTwingateResourceActiveState(resourceName string, expectedActiveState bool) sdk.TestCheckFunc {
	return CheckTwingateResource(resourceName, func(res *model.Resource) error {
		if res.IsActive != expectedActiveState {
			return fmt.Errorf("expected active state %v, got %v", expectedActiveState, res.IsActive) //nolint:err113
		}

		return nil
	})
}

type checkResourceActiveState struct {
	resourceAddress     string
	expectedActiveState bool
}

// CheckPlan implements the plan check logic.
func (e checkResourceActiveState) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	var resourceID string

	for _, rc := range req.Plan.ResourceChanges {
		if e.resourceAddress != rc.Address {
			continue
		}

		result, err := tfjsonpath.Traverse(rc.Change.Before, tfjsonpath.New(attr.ID))
		if err != nil {
			resp.Error = err

			return
		}

		resultID, ok := result.(string)
		if !ok {
			resp.Error = ErrInvalidPath

			return
		}

		resourceID = resultID

		break
	}

	if resourceID == "" {
		resp.Error = fmt.Errorf("%s - Resource not found in plan ResourceChanges", e.resourceAddress) //nolint:err113

		return
	}

	res, err := providerClient.ReadResource(ctx, resourceID)
	if err != nil {
		resp.Error = fmt.Errorf("failed to read resource: %w", err)

		return
	}

	if res.IsActive != e.expectedActiveState {
		resp.Error = fmt.Errorf("expected active state %v, got %v", e.expectedActiveState, res.IsActive) //nolint:err113

		return
	}
}

func CheckResourceActiveState(resourceAddress string, activeState bool) plancheck.PlanCheck {
	return checkResourceActiveState{
		resourceAddress:     resourceAddress,
		expectedActiveState: activeState,
	}
}

func CheckImportState(attributes map[string]string) func(data []*terraform.InstanceState) error {
	return func(data []*terraform.InstanceState) error {
		if len(data) != 1 {
			return fmt.Errorf("expected 1 resource, got %d", len(data)) //nolint:err113
		}

		res := data[0]
		for name, expected := range attributes {
			if res.Attributes[name] != expected {
				return fmt.Errorf("attribute %s doesn't match, expected: %s, got: %s", name, expected, res.Attributes[name]) //nolint:err113
			}
		}

		return nil
	}
}

func CheckTwingateRemoteNetworkDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateRemoteNetwork {
			continue
		}

		remoteNetworkID := rs.Primary.ID

		remoteNetwork, _ := providerClient.ReadRemoteNetworkByID(context.Background(), remoteNetworkID)
		if remoteNetwork != nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, remoteNetworkID)
		}
	}

	return nil
}

func CheckTwingateGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateGroup {
			continue
		}

		groupID := rs.Primary.ID

		group, _ := providerClient.ReadGroup(context.Background(), groupID)
		if group != nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, groupID)
		}
	}

	return nil
}

func CheckTwingateConnectorDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateConnector {
			continue
		}

		connectorID := rs.Primary.ID

		connector, _ := providerClient.ReadConnector(context.Background(), connectorID)
		if connector != nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, connectorID)
		}
	}

	return nil
}

func CheckTwingateDNSProfileDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateDNSFilteringProfile {
			continue
		}

		profileID := rs.Primary.ID

		profile, _ := providerClient.ReadDNSFilteringProfile(context.Background(), profileID)
		if profile != nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, profileID)
		}
	}

	return nil
}

func RevokeTwingateServiceKey(resourceName string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%w: %s", ErrResourceNotFound, resourceName)
		}

		resourceID := resourceState.Primary.ID
		if resourceID == "" {
			return ErrResourceIDNotSet
		}

		err := providerClient.RevokeServiceKey(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("failed to revoke service account key with ID %s: %w", resourceID, err)
		}

		return nil
	}
}

func CheckTwingateServiceKeyStatus(resourceName string, expectedStatus string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%w: %s", ErrResourceNotFound, resourceName)
		}

		if resourceState.Primary.ID == "" {
			return ErrResourceIDNotSet
		}

		serviceAccountKey, err := providerClient.ReadServiceKey(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read service account key with ID %s: %w", resourceState.Primary.ID, err)
		}

		if serviceAccountKey.Status != expectedStatus {
			return fmt.Errorf("expected status %v, got %v", expectedStatus, serviceAccountKey.Status) //nolint:err113
		}

		return nil
	}
}

func ListSecurityPolicies() ([]*model.SecurityPolicy, error) {
	if providerClient == nil {
		return nil, ErrClientNotInitialized
	}

	securityPolicies, err := providerClient.ReadSecurityPolicies(context.Background(), "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all security policies: %w", err)
	}

	if len(securityPolicies) == 0 {
		return nil, ErrSecurityPoliciesNotFound
	}

	return securityPolicies, nil
}

func AddResourceGroup(resourceName, groupName string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		groupID, err := getResourceID(state, groupName)
		if err != nil {
			return err
		}

		err = providerClient.AddResourceAccess(context.Background(), resourceID, []client.AccessInput{
			{PrincipalID: groupID},
		})
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to add group with ID %s: %w", resourceID, groupID, err)
		}

		return nil
	}
}

func DeleteResourceGroup(resourceName, groupName string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		groupID, err := getResourceID(state, groupName)
		if err != nil {
			return err
		}

		err = providerClient.RemoveResourceAccess(context.Background(), resourceID, []string{groupID})
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to delete group with ID %s: %w", resourceID, groupID, err)
		}

		return nil
	}
}

func CheckResourceGroupsLen(resourceName string, expectedGroupsLen int) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		resource, err := providerClient.ReadResource(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to read: %w", resourceID, err)
		}

		if len(resource.GroupsAccess) != expectedGroupsLen {
			return ErrGroupsLenMismatch(expectedGroupsLen, len(resource.GroupsAccess))
		}

		return nil
	}
}

func getResourceID(s *terraform.State, resourceName string) (string, error) {
	resourceState, ok := s.RootModule().Resources[resourceName]

	if !ok {
		return "", fmt.Errorf("%w: %s", ErrResourceNotFound, resourceName)
	}

	resourceID := resourceState.Primary.ID

	if resourceID == "" {
		return "", ErrResourceIDNotSet
	}

	return resourceID, nil
}

func AddResourceServiceAccount(resourceName, serviceAccountName string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		serviceAccountID, err := getResourceID(state, serviceAccountName)
		if err != nil {
			return err
		}

		err = providerClient.AddResourceAccess(context.Background(), resourceID, []client.AccessInput{
			{PrincipalID: serviceAccountID},
		})
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to add service account with ID %s: %w", resourceID, serviceAccountID, err)
		}

		return nil
	}
}

func AddDNSProfileAllowedDomains(resourceName string, domains []string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		profileID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		profile, err := providerClient.ReadDNSFilteringProfile(context.Background(), profileID)
		if err != nil {
			return fmt.Errorf("failed to fetch DNS profile with ID %s: %w", profileID, err)
		}

		profile.AllowedDomains = domains

		_, err = providerClient.UpdateDNSFilteringProfile(context.Background(), profile)
		if err != nil {
			return fmt.Errorf("DNS profile with ID %s failed to set new domains: %w", profileID, err)
		}

		return nil
	}
}

func DeleteResourceServiceAccount(resourceName, serviceAccountName string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		serviceAccountID, err := getResourceID(state, serviceAccountName)
		if err != nil {
			return err
		}

		err = providerClient.RemoveResourceAccess(context.Background(), resourceID, []string{serviceAccountID})
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to delete service account with ID %s: %w", resourceID, serviceAccountID, err)
		}

		return nil
	}
}

func CheckResourceServiceAccountsLen(resourceName string, expectedServiceAccountsLen int) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		resource, err := providerClient.ReadResource(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to read: %w", resourceID, err)
		}

		if len(resource.ServiceAccounts) != expectedServiceAccountsLen {
			return ErrServiceAccountsLenMismatch(expectedServiceAccountsLen, len(resource.ServiceAccounts))
		}

		return nil
	}
}

func CheckDNSProfileAllowedDomainsLen(resourceName string, expectedLen int) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		profile, err := providerClient.ReadDNSFilteringProfile(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("profile with ID %s failed to read: %w", resourceID, err)
		}

		if len(profile.AllowedDomains) != expectedLen {
			return ErrDNSProfileAllowedDomainsLenMismatch(expectedLen, len(profile.AllowedDomains))
		}

		return nil
	}
}

func CheckResourceSecurityPolicy(resourceName string, expectedSecurityPolicyID string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		resource, err := providerClient.ReadResource(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to read: %w", resourceID, err)
		}

		if resource.SecurityPolicyID != nil && *resource.SecurityPolicyID != expectedSecurityPolicyID {
			return fmt.Errorf("expected security_policy_id %s, got %s", expectedSecurityPolicyID, *resource.SecurityPolicyID) //nolint
		}

		return nil
	}
}

func CheckConnectorName(resourceName string, expectedName string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%w: %s", ErrResourceNotFound, resourceName)
		}

		if resourceState.Primary.ID == "" {
			return ErrResourceIDNotSet
		}

		connector, err := providerClient.ReadConnector(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read connector: %w", err)
		}

		if connector.Name != expectedName {
			return fmt.Errorf("expected name %v, got %v", expectedName, connector.Name) //nolint:err113
		}

		return nil
	}
}

func UpdateResourceSecurityPolicy(resourceName, securityPolicyID string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		resource, err := providerClient.ReadResource(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to read: %w", resourceID, err)
		}

		resource.SecurityPolicyID = &securityPolicyID

		_, err = providerClient.UpdateResource(context.Background(), resource)
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to update security_policy: %w", resourceID, err)
		}

		return nil
	}
}

func AddGroupUser(groupResource, groupName, terraformUserID string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		userID, err := getResourceID(state, getResourceNameFromID(terraformUserID))
		if err != nil {
			return err
		}

		resourceID, err := getResourceID(state, groupResource)
		if err != nil {
			return err
		}

		_, err = providerClient.UpdateGroup(context.Background(), &model.Group{
			ID:    resourceID,
			Name:  groupName,
			Users: []string{userID},
		})
		if err != nil {
			return fmt.Errorf("group with ID %s failed to add user with ID %s: %w", resourceID, userID, err)
		}

		return nil
	}
}

func getResourceNameFromID(terraformID string) string {
	return strings.TrimSuffix(terraformID, ".id")
}

func DeleteGroupUser(groupResource, terraformUserID string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		userID, err := getResourceID(state, getResourceNameFromID(terraformUserID))
		if err != nil {
			return err
		}

		groupID, err := getResourceID(state, groupResource)
		if err != nil {
			return err
		}

		err = providerClient.DeleteGroupUsers(context.Background(), groupID, []string{userID})
		if err != nil {
			return fmt.Errorf("group with ID %s failed to delete user with ID %s: %w", groupID, userID, err)
		}

		return nil
	}
}

func CheckGroupUsersLen(resourceName string, expectedUsersLen int) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		groupID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		group, err := providerClient.ReadGroup(context.Background(), groupID)
		if err != nil {
			return fmt.Errorf("group with ID %s failed to read: %w", groupID, err)
		}

		if len(group.Users) != expectedUsersLen {
			return ErrUsersLenMismatch(expectedUsersLen, len(group.Users))
		}

		return nil
	}
}

func GetTestUsers() ([]*model.User, error) {
	if providerClient == nil {
		return nil, ErrClientNotInitialized
	}

	users, err := providerClient.ReadUsers(context.Background(), nil)
	if err != nil {
		return nil, err //nolint
	}

	if len(users) == 0 {
		return nil, ErrResourceNotFound
	}

	return users, nil
}

func CheckTwingateUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateUser {
			continue
		}

		userID := rs.Primary.ID

		user, _ := providerClient.ReadUser(context.Background(), userID)
		if user != nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, userID)
		}
	}

	return nil
}

func CheckTwingateConnectorTokensInvalidated(s *terraform.State) error {
	for _, res := range s.RootModule().Resources {
		if res.Type != resource.TwingateConnectorTokens {
			continue
		}

		connectorID := res.Primary.ID
		accessToken := res.Primary.Attributes[attr.AccessToken]
		refreshToken := res.Primary.Attributes[attr.RefreshToken]

		err := providerClient.VerifyConnectorTokens(context.Background(), refreshToken, accessToken)
		// expecting error here, since tokens invalidated
		if err == nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, connectorID)
		}
	}

	return nil
}

func GetTestUser() (*model.User, error) {
	if providerClient == nil {
		return nil, ErrClientNotInitialized
	}

	users, err := providerClient.ReadUsers(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get test users: %w", err)
	}

	if len(users) == 0 {
		return nil, ErrResourceNotFound
	}

	return users[0], nil
}

func CheckTwingateConnectorAndRemoteNetworkDestroy(s *terraform.State) error {
	if err := CheckTwingateConnectorDestroy(s); err != nil {
		return err
	}

	return CheckTwingateRemoteNetworkDestroy(s)
}
