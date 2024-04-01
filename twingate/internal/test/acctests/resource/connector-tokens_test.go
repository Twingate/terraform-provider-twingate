package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRemoteConnectorWithTokens(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector With Tokens", func(t *testing.T) {
		const terraformResourceName = "test_t1"
		theResource := acctests.TerraformConnectorTokens(terraformResourceName)
		remoteNetworkName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateConnectorTokensInvalidated,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateConnectorTokens(terraformResourceName, remoteNetworkName),
					Check: acctests.ComposeTestCheckFunc(
						checkTwingateConnectorTokensSet(theResource),
					),
				},
			},
		})
	})
}

func terraformResourceTwingateConnectorTokens(terraformResourceName, remoteNetworkName string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_connector_tokens" "%s" {
	  connector_id = twingate_connector.%s.id
      keepers = {
         foo = "bar"
      }
	}
	`, terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName), terraformResourceName, terraformResourceName)
}

func checkTwingateConnectorTokensSet(connectorNameTokens string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		connectorTokens, ok := s.RootModule().Resources[connectorNameTokens]

		if !ok {
			return fmt.Errorf("not found: %s", connectorNameTokens)
		}

		if connectorTokens.Primary.ID == "" {
			return fmt.Errorf("no connectorTokensID set")
		}

		if connectorTokens.Primary.Attributes[attr.AccessToken] == "" {
			return fmt.Errorf("no access token set")
		}

		if connectorTokens.Primary.Attributes[attr.RefreshToken] == "" {
			return fmt.Errorf("no refresh token set")
		}

		return nil
	}
}
