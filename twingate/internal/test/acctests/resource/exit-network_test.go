package resource

import (
	"fmt"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTwingateExitNetworkCreate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test000"
	theResource := acctests.TerraformExitNetwork(terraformResourceName)
	networkName := test.RandomName()
	networkLocation := model.LocationAzure

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateExitNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createExitNetworkWithLocation(terraformResourceName, networkName, networkLocation),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, networkName),
					sdk.TestCheckResourceAttr(theResource, attr.Location, networkLocation),
				),
			},
		},
	})
}

func createExitNetworkWithLocation(terraformResourceName, name, location string) string {
	return fmt.Sprintf(`
	resource "twingate_exit_network" "%s" {
	  name = "%s"
	  location = "%s"
	}
	`, terraformResourceName, name, location)
}

func TestAccTwingateExitNetworkUpdate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test001"
	theResource := acctests.TerraformExitNetwork(terraformResourceName)
	nameBefore := test.RandomName()
	nameAfter := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateExitNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceExitNetwork(terraformResourceName, nameBefore),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, nameBefore),
					sdk.TestCheckResourceAttr(theResource, attr.Location, model.LocationOther),
				),
			},
			{
				Config: createExitNetworkWithLocation(terraformResourceName, nameAfter, model.LocationAWS),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, nameAfter),
					sdk.TestCheckResourceAttr(theResource, attr.Location, model.LocationAWS),
				),
			},
		},
	})
}

func terraformResourceExitNetwork(terraformResourceName, name string) string {
	return fmt.Sprintf(`
	resource "twingate_exit_network" "%s" {
	  name = "%s"
	}
	`, terraformResourceName, name)
}

func TestAccTwingateExitNetworkDeleteNonExisting(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test002"
	theResource := acctests.TerraformExitNetwork(terraformResourceName)
	networkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateExitNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  terraformResourceExitNetwork(terraformResourceName, networkName),
				Destroy: true,
			},
			{
				Config: terraformResourceExitNetwork(terraformResourceName, networkName),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateExitNetworkReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test003"
	theResource := acctests.TerraformExitNetwork(terraformResourceName)
	exitNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateExitNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceExitNetwork(terraformResourceName, exitNetworkName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateExitNetwork),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceExitNetwork(terraformResourceName, exitNetworkName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func TestAccTwingateExitNetworkUpdateWithTheSameName(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test004"
	theResource := acctests.TerraformExitNetwork(terraformResourceName)
	name := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateExitNetworkDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceExitNetwork(terraformResourceName, name),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
					sdk.TestCheckResourceAttr(theResource, attr.Location, model.LocationOther),
				),
			},
			{
				Config: createExitNetworkWithLocation(terraformResourceName, name, model.LocationAWS),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
					sdk.TestCheckResourceAttr(theResource, attr.Location, model.LocationAWS),
				),
			},
		},
	})
}
