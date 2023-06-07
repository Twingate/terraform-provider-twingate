package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
						acctests.CheckTwingateConnectorTokensSet(theResource),
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

func TestAccRemoteConnectorRecreation(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector Recreation", func(t *testing.T) {
		const terraformResourceName = "test_t2"
		theResource := acctests.TerraformConnectorTokens(terraformResourceName)
		remoteNetworkName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateConnectorTokensInvalidated,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateConnectorTokensWithKeeper(terraformResourceName, remoteNetworkName, test.RandomName()),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateConnectorTokensSet(theResource),
					),
				},
				{
					Config: terraformResourceTwingateConnectorTokensWithKeeper(terraformResourceName, remoteNetworkName, test.RandomName()),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateConnectorTokensSet(theResource),
					),
				},
			},
		})
	})
}

func terraformResourceTwingateConnectorTokensWithKeeper(terraformResourceName, remoteNetworkName, keeper string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_connector_tokens" "%s" {
	  connector_id = twingate_connector.%s.id
     keepers = {
        foo = "%s"
     }
	}
	`, terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName), terraformResourceName, terraformResourceName, keeper)
}
