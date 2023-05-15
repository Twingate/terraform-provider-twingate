package acctests

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	ErrResourceIDNotSet         = errors.New("id not set")
	ErrResourceNotFound         = errors.New("resource not found")
	ErrResourceStillPresent     = errors.New("resource still present")
	ErrResourceFoundInState     = errors.New("this resource should not be here")
	ErrUnknownResourceType      = errors.New("unknown resource type")
	ErrClientNotInited          = errors.New("meta client not inited")
	ErrSecurityPoliciesNotFound = errors.New("security policies not found")
)

func ErrServiceAccountsLenMismatch(expected, actual int) error {
	return fmt.Errorf("expected %d service accounts, actual - %d", expected, actual) //nolint
}

func ErrGroupsLenMismatch(expected, actual int) error {
	return fmt.Errorf("expected %d groups, actual - %d", expected, actual) //nolint
}

func ErrUsersLenMismatch(expected, actual int) error {
	return fmt.Errorf("expected %d users, actual - %d", expected, actual) //nolint
}

var Provider *schema.Provider                                     //nolint:gochecknoglobals
var ProviderFactories map[string]func() (*schema.Provider, error) //nolint:gochecknoglobals

//nolint:gochecknoinits
func init() {
	Provider = twingate.Provider("test")

	ProviderFactories = map[string]func() (*schema.Provider, error){
		"twingate": func() (*schema.Provider, error) {
			return Provider, nil
		},
	}
}

func SetPageLimit(limit int) {
	if err := os.Setenv(client.EnvPageLimit, fmt.Sprintf("%d", limit)); err != nil {
		log.Fatal("failed to set page limit", err)
	}
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
	providerClient := Provider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateServiceAccount {
			continue
		}

		serviceAccountID := rs.Primary.ID

		_, err := providerClient.ReadShallowServiceAccount(context.Background(), serviceAccountID)
		if err == nil {
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

	providerClient := Provider.Meta().(*client.Client)

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
	providerClient := Provider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateResource {
			continue
		}

		resourceID := rs.Primary.ID

		err := providerClient.DeleteResource(context.Background(), resourceID)
		// expecting error here , since the resource is already gone
		if err == nil {
			return fmt.Errorf("%w: with ID %s", ErrResourceStillPresent, resourceID)
		}
	}

	return nil
}

func DeactivateTwingateResource(resourceName string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

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

func CheckTwingateResourceActiveState(resourceName string, expectedActiveState bool) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

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

		if res.IsActive != expectedActiveState {
			return fmt.Errorf("expected active state %v, got %v", expectedActiveState, res.IsActive) //nolint:goerr113
		}

		return nil
	}
}

func CheckImportState(attributes map[string]string) func(data []*terraform.InstanceState) error {
	return func(data []*terraform.InstanceState) error {
		if len(data) != 1 {
			return fmt.Errorf("expected 1 resource, got %d", len(data)) //nolint:goerr113
		}

		res := data[0]
		for name, expected := range attributes {
			if res.Attributes[name] != expected {
				return fmt.Errorf("attribute %s doesn't match, expected: %s, got: %s", name, expected, res.Attributes[name]) //nolint:goerr113
			}
		}

		return nil
	}
}

func CheckTwingateRemoteNetworkDestroy(s *terraform.State) error {
	providerClient := Provider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateRemoteNetwork {
			continue
		}

		remoteNetworkID := rs.Primary.ID

		err := providerClient.DeleteRemoteNetwork(context.Background(), remoteNetworkID)
		// expecting error here, since the network is already gone
		if err == nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, remoteNetworkID)
		}
	}

	return nil
}

func CheckTwingateGroupDestroy(s *terraform.State) error {
	providerClient := Provider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateGroup {
			continue
		}

		groupID := rs.Primary.ID

		err := providerClient.DeleteGroup(context.Background(), groupID)
		if err == nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, groupID)
		}
	}

	return nil
}

func CheckTwingateConnectorDestroy(s *terraform.State) error {
	providerClient := Provider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateConnector {
			continue
		}

		connectorID := rs.Primary.ID

		err := providerClient.DeleteConnector(context.Background(), connectorID)
		// expecting error here, since the network is already gone
		if err == nil {
			return fmt.Errorf("%w with ID %s", ErrResourceStillPresent, connectorID)
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

		client := Provider.Meta().(*client.Client)

		err := client.RevokeServiceKey(context.Background(), resourceID)
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

		client := Provider.Meta().(*client.Client)

		serviceAccountKey, err := client.ReadServiceKey(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read service account key with ID %s: %w", resourceState.Primary.ID, err)
		}

		if serviceAccountKey.Status != expectedStatus {
			return fmt.Errorf("expected status %v, got %v", expectedStatus, serviceAccountKey.Status) //nolint:goerr113
		}

		return nil
	}
}

func ListSecurityPolicies() ([]*model.SecurityPolicy, error) {
	if Provider.Meta() == nil {
		return nil, ErrClientNotInited
	}

	client := Provider.Meta().(*client.Client)

	securityPolicies, err := client.ReadSecurityPolicies(context.Background())
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
		providerClient := Provider.Meta().(*client.Client)

		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		groupID, err := getResourceID(state, groupName)
		if err != nil {
			return err
		}

		err = providerClient.AddResourceGroups(context.Background(), &model.Resource{
			ID:     resourceID,
			Groups: []string{groupID},
		})
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to add group with ID %s: %w", resourceID, groupID, err)
		}

		return nil
	}
}

func DeleteResourceGroup(resourceName, groupName string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		groupID, err := getResourceID(state, groupName)
		if err != nil {
			return err
		}

		err = providerClient.DeleteResourceGroups(context.Background(), resourceID, []string{groupID})
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to delete group with ID %s: %w", resourceID, groupID, err)
		}

		return nil
	}
}

func CheckResourceGroupsLen(resourceName string, expectedGroupsLen int) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		resource, err := providerClient.ReadResource(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to read: %w", resourceID, err)
		}

		if len(resource.Groups) != expectedGroupsLen {
			return ErrGroupsLenMismatch(expectedGroupsLen, len(resource.Groups))
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
		providerClient := Provider.Meta().(*client.Client)

		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		serviceAccountID, err := getResourceID(state, serviceAccountName)
		if err != nil {
			return err
		}

		_, err = providerClient.UpdateServiceAccount(context.Background(), &model.ServiceAccount{
			ID:        serviceAccountID,
			Resources: []string{resourceID},
		})
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to add service account with ID %s: %w", resourceID, serviceAccountID, err)
		}

		return nil
	}
}

func DeleteResourceServiceAccount(resourceName, serviceAccountName string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		serviceAccountID, err := getResourceID(state, serviceAccountName)
		if err != nil {
			return err
		}

		err = providerClient.DeleteResourceServiceAccounts(context.Background(), resourceID, []string{serviceAccountID})
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to delete service account with ID %s: %w", resourceID, serviceAccountID, err)
		}

		return nil
	}
}

func CheckResourceServiceAccountsLen(resourceName string, expectedServiceAccountsLen int) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

		resourceID, err := getResourceID(state, resourceName)
		if err != nil {
			return err
		}

		resource, err := providerClient.ReadResource(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("resource with ID %s failed to read: %w", resourceID, err)
		}

		serviceAccounts, err := providerClient.ReadServiceAccounts(context.Background())
		if err != nil {
			return fmt.Errorf("failed to read service accounts: %w", err)
		}

		serviceAccountIDs := make(map[string]bool)

		for _, account := range serviceAccounts {
			if utils.Contains(account.Resources, resource.ID) {
				serviceAccountIDs[account.ID] = true
			}
		}

		resource.ServiceAccounts = utils.MapKeys(serviceAccountIDs)

		if len(resource.ServiceAccounts) != expectedServiceAccountsLen {
			return ErrServiceAccountsLenMismatch(expectedServiceAccountsLen, len(resource.ServiceAccounts))
		}

		return nil
	}
}

func AddGroupUser(groupResource, groupName, userID string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

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

func DeleteGroupUser(groupResource, userID string) sdk.TestCheckFunc {
	return func(state *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

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
		providerClient := Provider.Meta().(*client.Client)

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
	if Provider.Meta() == nil {
		return nil, ErrClientNotInited
	}

	client := Provider.Meta().(*client.Client)

	users, err := client.ReadUsers(context.Background())
	if err != nil {
		return nil, err //nolint
	}

	if len(users) == 0 {
		return nil, ErrResourceNotFound
	}

	return users, nil
}

func CheckTwingateUserDestroy(s *terraform.State) error {
	providerClient := Provider.Meta().(*client.Client)

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
