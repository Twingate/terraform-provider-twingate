package twingate

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRemoteConnectorWithTokens(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector With Tokens", func(t *testing.T) {

		const connectorTokensResource = "twingate_connector_tokens.test_t1"
		remoteNetworkName := getRandomName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorTokensInvalidated,
			Steps: []resource.TestStep{
				{
					Config: createConnectorTokensWithKeepers(remoteNetworkName),
					Check: ComposeTestCheckFunc(
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
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_connector_tokens" {
			continue
		}

		connectorId := rs.Primary.ID
		accessToken := rs.Primary.Attributes["access_token"]
		refreshToken := rs.Primary.Attributes["refresh_token"]

		err := client.verifyConnectorTokens(context.Background(), refreshToken, accessToken)
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
