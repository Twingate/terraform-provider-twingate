package twingate

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRemoteConnector_withKeys(t *testing.T) {

	remoteNetworkName := acctest.RandomWithPrefix("tf-acc")
	connectorKeysResource := "twingate_connector_tokens.test"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateConnectorTokensInvalidated,
		Steps: []resource.TestStep{
			{
				Config: testTwingateConnectorKeys(remoteNetworkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateConnectorKeysExists(connectorKeysResource),
				),
			},
		},
	})
}

func testTwingateConnectorKeys(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_connector" "test" {
	  remote_network_id = twingate_remote_network.test.id
	}
	resource "twingate_connector_tokens" "test" {
	  connector_id = twingate_connector.test.id
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
		err := client.verifyConnectorTokens(&refreshToken, &accessToken)
		// expecting error here , Since tokens invalidated
		if err == nil {
			return fmt.Errorf("connector with ID %s tokens that should be inactive are still active: ", connectorId)
		}
	}

	return nil
}

func testAccCheckTwingateConnectorKeysExists(connectorNameKeys string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		connectorKeys, ok := s.RootModule().Resources[connectorNameKeys]

		if !ok {
			return fmt.Errorf("Not found: %s ", connectorNameKeys)
		}

		if connectorKeys.Primary.ID == "" {
			return fmt.Errorf("No connectorKeysID set ")
		}

		if connectorKeys.Primary.Attributes["access_token"] == "" {
			return fmt.Errorf("No access token set ")
		}

		if connectorKeys.Primary.Attributes["refresh_token"] == "" {
			return fmt.Errorf("No refresh token set ")
		}

		return nil
	}
}
