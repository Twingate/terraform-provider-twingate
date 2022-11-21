package resource

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testRegexp = regexp.MustCompile(test.Prefix() + ".*")

func TestAccRemoteConnectorCreate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector", func(t *testing.T) {
		const terraformResourceName = "test_c1"
		theResource := acctests.TerraformConnector(terraformResourceName)
		remoteNetworkName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateConnectorDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName),
					Check: acctests.ComposeTestCheckFunc(
						checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
						sdk.TestCheckResourceAttrSet(theResource, nameAttr),
					),
				},
			},
		})
	})
}

func TestAccRemoteConnectorWithCustomName(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector With Custom Name", func(t *testing.T) {
		const terraformResourceName = "test_c2"
		theResource := acctests.TerraformConnector(terraformResourceName)
		remoteNetworkName := test.RandomName()
		connectorName := test.RandomConnectorName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateConnectorDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, connectorName),
					Check: acctests.ComposeTestCheckFunc(
						checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
						sdk.TestMatchResourceAttr(theResource, nameAttr, regexp.MustCompile(connectorName)),
					),
				},
			},
		})
	})
}

func TestAccRemoteConnectorImport(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector - Import", func(t *testing.T) {
		const terraformResourceName = "test_c3"
		theResource := acctests.TerraformConnector(terraformResourceName)
		remoteNetworkName := test.RandomName()
		connectorName := test.RandomConnectorName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateConnectorDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, connectorName),
					Check: acctests.ComposeTestCheckFunc(
						checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
						sdk.TestMatchResourceAttr(theResource, nameAttr, testRegexp),
					),
				},
				{
					ResourceName:      theResource,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccRemoteConnectorNotAllowedToChangeRemoteNetworkId(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector - should fail on remote_network_id update", func(t *testing.T) {
		const (
			terraformConnectorName      = "test_c4"
			terraformRemoteNetworkName1 = "test_c4_1"
			terraformRemoteNetworkName2 = "test_c4_2"
		)
		theResource := acctests.TerraformConnector(terraformConnectorName)
		remoteNetworkName1 := test.RandomName()
		remoteNetworkName2 := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateConnectorDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateConnector(terraformRemoteNetworkName1, terraformConnectorName, remoteNetworkName1),
					Check: acctests.ComposeTestCheckFunc(
						checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformRemoteNetworkName1)),
					),
				},
				{
					Config:      terraformResourceTwingateConnector(terraformRemoteNetworkName2, terraformConnectorName, remoteNetworkName2),
					ExpectError: regexp.MustCompile(resource.ErrNotAllowChangeRemoteNetworkID.Error()),
				},
			},
		})
	})
}

func TestAccTwingateConnectorReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector ReCreate After Deletion", func(t *testing.T) {
		const terraformResourceName = "test_c5"
		theResource := acctests.TerraformConnector(terraformResourceName)
		remoteNetworkName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateConnectorDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName),
					Check: acctests.ComposeTestCheckFunc(
						checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
						acctests.DeleteTwingateResource(theResource, resource.TwingateConnector),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName),
					Check: acctests.ComposeTestCheckFunc(
						checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					),
				},
			},
		})
	})
}

func terraformResourceTwingateConnector(terraformRemoteNetworkName, terraformConnectorName, remoteNetworkName string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_connector" "%s" {
	  remote_network_id = twingate_remote_network.%s.id
	}
	`, terraformResourceRemoteNetwork(terraformRemoteNetworkName, remoteNetworkName), terraformConnectorName, terraformRemoteNetworkName)
}

func terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, connectorName string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_connector" "%s" {
	  remote_network_id = twingate_remote_network.%s.id
      name  = "%s"
	}
	`, terraformResourceRemoteNetwork(terraformResourceName, remoteNetworkName), terraformResourceName, terraformResourceName, connectorName)
}

func checkTwingateConnectorSetWithRemoteNetwork(connectorResource, remoteNetworkResource string) sdk.TestCheckFunc {
	return func(s *terraform.State) error {
		connector, ok := s.RootModule().Resources[connectorResource]
		if !ok {
			return fmt.Errorf("Not found: %s ", connectorResource)
		}

		if connector.Primary.ID == "" {
			return fmt.Errorf("No connectorID set ")
		}

		remoteNetwork, ok := s.RootModule().Resources[remoteNetworkResource]
		if !ok {
			return fmt.Errorf("Not found: %s ", remoteNetworkResource)
		}

		if connector.Primary.Attributes["remote_network_id"] != remoteNetwork.Primary.ID {
			return fmt.Errorf("Remote Network ID not set properly in the connector ")
		}

		return nil
	}
}
