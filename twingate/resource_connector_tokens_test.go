package twingate

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRemoteConnector_withTokens(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector With Tokens", func(t *testing.T) {

		remoteNetworkName := acctest.RandomWithPrefix("tf-acc")
		connectorTokensResource := "twingate_connector_tokens.test"

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorTokensInvalidated,
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnectorTokensWithKeepers(remoteNetworkName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorTokensExists(connectorTokensResource),
					),
				},
			},
		})
	})
}

func testTwingateConnectorTokensWithKeepers(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_connector" "test" {
	  remote_network_id = twingate_remote_network.test.id
	}
	resource "twingate_connector_tokens" "test" {
	  connector_id = twingate_connector.test.id
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
