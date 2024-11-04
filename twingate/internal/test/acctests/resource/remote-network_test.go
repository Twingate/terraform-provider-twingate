package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTwingateRemoteNetworkCreate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test000"
	theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
	networkName := test.RandomName()
	networkLocation := model.LocationAzure

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createRemoteNetworkWithLocation(terraformResourceName, networkName, networkLocation),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, networkName),
					sdk.TestCheckResourceAttr(theResource, attr.Location, networkLocation),
				),
			},
		},
	})
}

func createRemoteNetworkWithLocation(terraformResourceName, name, location string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	  location = "%s"
	}
	`, terraformResourceName, name, location)
}

func TestAccTwingateRemoteNetworkUpdate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test001"
	theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
	nameBefore := test.RandomName()
	nameAfter := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceRemoteNetwork(terraformResourceName, nameBefore),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, nameBefore),
					sdk.TestCheckResourceAttr(theResource, attr.Location, model.LocationOther),
				),
			},
			{
				Config: createRemoteNetworkWithLocation(terraformResourceName, nameAfter, model.LocationAWS),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, nameAfter),
					sdk.TestCheckResourceAttr(theResource, attr.Location, model.LocationAWS),
				),
			},
		},
	})
}

func terraformResourceRemoteNetwork(terraformResourceName, name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	`, terraformResourceName, name)
}

func TestAccTwingateRemoteNetworkDeleteNonExisting(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test002"
	theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
	networkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  terraformResourceRemoteNetwork(terraformResourceName, networkName),
				Destroy: true,
			},
			{
				Config: terraformResourceRemoteNetwork(terraformResourceName, networkName),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateRemoteNetworkReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test003"
	theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
	remoteNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceRemoteNetwork(terraformResourceName, remoteNetworkName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateRemoteNetwork),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceRemoteNetwork(terraformResourceName, remoteNetworkName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func TestAccTwingateRemoteNetworkUpdateWithTheSameName(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test004"
	theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
	name := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateRemoteNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceRemoteNetwork(terraformResourceName, name),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
					sdk.TestCheckResourceAttr(theResource, attr.Location, model.LocationOther),
				),
			},
			{
				Config: createRemoteNetworkWithLocation(terraformResourceName, name, model.LocationAWS),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
					sdk.TestCheckResourceAttr(theResource, attr.Location, model.LocationAWS),
				),
			},
		},
	})
}
