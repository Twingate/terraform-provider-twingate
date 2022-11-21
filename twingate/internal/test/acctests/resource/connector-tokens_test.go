package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRemoteConnectorWithTokens(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector With Tokens", func(t *testing.T) {
		const terraformResourceName = "test_t1"
		theResource := acctests.TerraformConnectorTokens(terraformResourceName)
		remoteNetworkName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      checkTwingateConnectorTokensInvalidated,
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

func checkTwingateConnectorTokensInvalidated(s *terraform.State) error {
	c := acctests.Provider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resource.TwingateConnectorTokens {
			continue
		}

		connectorId := rs.Primary.ID
		accessToken := rs.Primary.Attributes["access_token"]
		refreshToken := rs.Primary.Attributes["refresh_token"]

		err := c.VerifyConnectorTokens(context.Background(), refreshToken, accessToken)
		// expecting error here , Since tokens invalidated
		if err == nil {
			return fmt.Errorf("connector with ID %s tokens that should be inactive are still active", connectorId)
		}
	}

	return nil
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

		if connectorTokens.Primary.Attributes["access_token"] == "" {
			return fmt.Errorf("no access token set")
		}

		if connectorTokens.Primary.Attributes["refresh_token"] == "" {
			return fmt.Errorf("no refresh token set")
		}

		return nil
	}
}
