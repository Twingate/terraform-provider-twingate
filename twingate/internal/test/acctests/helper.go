package acctests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	ErrResourceIDNotSet     = errors.New("id not set")
	ErrResourceNotFound     = errors.New("resource not found")
	ErrResourceFoundInState = errors.New("this resource should not be here")
	ErrResourceStillPresent = errors.New("resource still present")
)

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

func DeleteTwingateResource(resourceName, resourceType string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		providerClient := Provider.Meta().(*client.Client)

		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%w: %s ", ErrResourceNotFound, resourceName)
		}

		resourceID := resourceState.Primary.ID

		if resourceID == "" {
			return ErrResourceIDNotSet
		}

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
		default:
			return fmt.Errorf("%s unknown resource type", resourceType) //nolint:goerr113
		}

		if err != nil {
			return fmt.Errorf("%s with ID %s still active: %w", resourceType, resourceID, err)
		}

		return nil
	}
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
