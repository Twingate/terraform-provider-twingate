package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRemoteConnectorWithTokens(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Remote Connector With Tokens", func(t *testing.T) {

		const connectorTokensResource = "twingate_connector_tokens.test_t1"
		remoteNetworkName := test.RandomName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorTokensInvalidated,
			Steps: []resource.TestStep{
				{
					Config: createConnectorTokensWithKeepers(remoteNetworkName),
					Check: acctests.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorTokensExists(connectorTokensResource),
					),
				},
			},
		})
	})
}

func createConnectorTokensWithKeepers(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_t1" {
	  name = "%s"
	}
	resource "twingate_connector" "test_t1" {
	  remote_network_id = twingate_remote_network.test_t1.id
	}
	resource "twingate_connector_tokens" "test_t1" {
	  connector_id = twingate_connector.test_t1.id
      keepers = {
         foo = "bar"
      }
	}
	`, remoteNetworkName)
}

func testAccCheckTwingateConnectorTokensInvalidated(s *terraform.State) error {
	client := acctests.Provider.Meta().(*transport.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_connector_tokens" {
			continue
		}

		connectorId := rs.Primary.ID
		accessToken := rs.Primary.Attributes["access_token"]
		refreshToken := rs.Primary.Attributes["refresh_token"]

		err := client.VerifyConnectorTokens(context.Background(), refreshToken, accessToken)
		// expecting error here , Since tokens invalidated
		if err == nil {
			return fmt.Errorf("connector with ID %s tokens that should be inactive are still active", connectorId)
		}
	}

	return nil
}

func testAccCheckTwingateConnectorTokensExists(connectorNameTokens string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		connectorTokens, ok := s.RootModule().Resources[connectorNameTokens]

		if !ok {
			return fmt.Errorf("not found: %s", connectorNameTokens)
		}

		if connectorTokens.Primary.ID == "" {
			return fmt.Errorf("no connectorTokensID set")
		}

		if connectorTokens.Primary.Attributes["access_token"] == "" {
			return fmt.Errorf("no access token set")
		}

		if connectorTokens.Primary.Attributes["refresh_token"] == "" {
			return fmt.Errorf("no refresh token set")
		}

		return nil
	}
}
