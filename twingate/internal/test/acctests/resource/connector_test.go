package resource

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRemoteConnectorCreate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_c1"
	theResource := acctests.TerraformConnector(terraformResourceName)
	remoteNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName),
				Check: acctests.ComposeTestCheckFunc(
					checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					sdk.TestCheckResourceAttrSet(theResource, attr.Name),
					sdk.TestCheckResourceAttrSet(theResource, attr.State),
				),
			},
		},
	})
}

func TestAccRemoteConnectorUpdateNetworkName(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_n1"
	theResource := acctests.TerraformConnector(terraformResourceName)
	networkResource := acctests.TerraformRemoteNetwork(terraformResourceName)
	remoteNetworkName1 := test.RandomName()
	remoteNetworkName2 := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName1),
				Check: acctests.ComposeTestCheckFunc(
					checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					sdk.TestCheckResourceAttr(networkResource, attr.Name, remoteNetworkName1),
				),
			},
			{
				Config: terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName2),
				Check: acctests.ComposeTestCheckFunc(
					checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					sdk.TestCheckResourceAttr(networkResource, attr.Name, remoteNetworkName2),
				),
			},
		},
	})
}

func TestAccRemoteConnectorWithCustomName(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_c2"
	theResource := acctests.TerraformConnector(terraformResourceName)
	remoteNetworkName := test.RandomName()
	connectorName := "  Some        connector   name     "
	sanitizedName := "Some connector name"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					sdk.TestCheckResourceAttr(theResource, attr.Name, connectorName),
					acctests.CheckConnectorName(theResource, sanitizedName),
				),
			},
			{
				// expecting no changes
				PlanOnly: true,
				Config:   terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					sdk.TestCheckResourceAttr(theResource, attr.Name, connectorName),
					acctests.CheckConnectorName(theResource, sanitizedName),
				),
			},
		},
	})
}

func TestAccRemoteConnectorImport(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_c3"
	theResource := acctests.TerraformConnector(terraformResourceName)
	remoteNetworkName := test.RandomName()
	connectorName := test.RandomConnectorName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					sdk.TestMatchResourceAttr(theResource, attr.Name, regexp.MustCompile(connectorName[:len(connectorName)-3]+".*")),
				),
			},
			{
				ResourceName:      theResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRemoteConnectorNotAllowedToChangeRemoteNetworkId(t *testing.T) {
	t.Parallel()

	const (
		terraformConnectorName      = "test_c4"
		terraformRemoteNetworkName1 = "test_c4_1"
		terraformRemoteNetworkName2 = "test_c4_2"
	)
	theResource := acctests.TerraformConnector(terraformConnectorName)
	remoteNetworkName1 := test.RandomName()
	remoteNetworkName2 := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
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
}

func TestAccTwingateConnectorReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_c5"
	theResource := acctests.TerraformConnector(terraformResourceName)
	remoteNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
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

		if connector.Primary.Attributes[attr.RemoteNetworkID] != remoteNetwork.Primary.ID {
			return fmt.Errorf("Remote Network ID not set properly in the connector ")
		}

		return nil
	}
}

func TestAccRemoteConnectorUpdateName(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_c6"
	theResource := acctests.TerraformConnector(terraformResourceName)
	remoteNetworkName := test.RandomName()
	connectorName := test.RandomConnectorName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName),
				Check: acctests.ComposeTestCheckFunc(
					checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					sdk.TestCheckResourceAttrSet(theResource, attr.Name),
				),
			},
			{
				Config: terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, connectorName),
				),
			},
		},
	})
}

func TestAccRemoteConnectorCreateWithNotificationStatus(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_c7"
	theResource := acctests.TerraformConnector(terraformResourceName)
	remoteNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName),
				Check: acctests.ComposeTestCheckFunc(
					checkTwingateConnectorSetWithRemoteNetwork(theResource, acctests.TerraformRemoteNetwork(terraformResourceName)),
					sdk.TestCheckResourceAttrSet(theResource, attr.Name),
				),
			},
			{
				// expecting no changes, as by default notifications enabled
				PlanOnly: true,
				Config:   terraformResourceTwingateConnectorWithNotificationStatus(terraformResourceName, terraformResourceName, remoteNetworkName, true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.StatusUpdatesEnabled, "true"),
				),
			},
			{
				Config: terraformResourceTwingateConnectorWithNotificationStatus(terraformResourceName, terraformResourceName, remoteNetworkName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.StatusUpdatesEnabled, "false"),
				),
			},
			{
				// expecting no changes, when user removes `status_updates_enabled` field from terraform
				PlanOnly: true,
				Config:   terraformResourceTwingateConnector(terraformResourceName, terraformResourceName, remoteNetworkName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.StatusUpdatesEnabled, "false"),
				),
			},
		},
	})
}

func terraformResourceTwingateConnectorWithNotificationStatus(terraformRemoteNetworkName, terraformConnectorName, remoteNetworkName string, notificationStatus bool) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_connector" "%s" {
	  remote_network_id = twingate_remote_network.%s.id
	  status_updates_enabled = %v
	}
	`, terraformResourceRemoteNetwork(terraformRemoteNetworkName, remoteNetworkName), terraformConnectorName, terraformRemoteNetworkName, notificationStatus)
}

func TestAccRemoteConnectorCreateWithNotificationStatusFalse(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_c8"
	theResource := acctests.TerraformConnector(terraformResourceName)
	remoteNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateConnectorWithNotificationStatus(terraformResourceName, terraformResourceName, remoteNetworkName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.StatusUpdatesEnabled, "false"),
				),
			},
		},
	})
}

func TestAccRemoteConnectorNameMustBeAtLeast3CharactersLong(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test_c9"
	theResource := acctests.TerraformConnector(terraformResourceName)
	remoteNetworkName := test.RandomName()
	expectedError := regexp.MustCompile("Attribute name must be at least 3 characters long")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorAndRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, ""),
				ExpectError: expectedError,
			},
			{
				Config:      terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, "   ab    "),
				ExpectError: expectedError,
			},
			{
				Config: terraformResourceTwingateConnectorWithName(terraformResourceName, remoteNetworkName, "   a    b    "),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckConnectorName(theResource, "a b"),
				),
			},
		},
	})

}
