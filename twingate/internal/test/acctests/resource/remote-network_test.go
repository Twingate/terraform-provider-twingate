package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const locationAttr = "location"

func TestAccTwingateRemoteNetworkCreate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Network Create", func(t *testing.T) {
		const terraformResourceName = "test000"
		theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
		networkName := test.RandomName()
		networkLocation := model.LocationAzure

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateRemoteNetworkDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createRemoteNetworkWithLocation(terraformResourceName, networkName, networkLocation),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, networkName),
						sdk.TestCheckResourceAttr(theResource, locationAttr, networkLocation),
					),
				},
			},
		})
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
	t.Run("Test Twingate Resource : Acc Remote Network Update", func(t *testing.T) {
		const terraformResourceName = "test001"
		theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
		nameBefore := test.RandomName()
		nameAfter := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateRemoteNetworkDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceRemoteNetwork(terraformResourceName, nameBefore),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, nameBefore),
						sdk.TestCheckResourceAttr(theResource, locationAttr, model.LocationOther),
					),
				},
				{
					Config: createRemoteNetworkWithLocation(terraformResourceName, nameAfter, model.LocationAWS),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, nameAfter),
						sdk.TestCheckResourceAttr(theResource, locationAttr, model.LocationAWS),
					),
				},
			},
		})
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
	t.Run("Test Twingate Resource : Acc Remote Network Delete NonExisting", func(t *testing.T) {
		const terraformResourceName = "test002"
		theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
		remoteNetworkNameBefore := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateRemoteNetworkDestroy,
			Steps: []sdk.TestStep{
				{
					Config:  terraformResourceRemoteNetwork(terraformResourceName, remoteNetworkNameBefore),
					Destroy: true,
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceDoesNotExists(theResource),
					),
				},
			},
		})
	})
}

func TestAccTwingateRemoteNetworkReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Network Re Create After Deletion", func(t *testing.T) {
		const terraformResourceName = "test003"
		theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
		remoteNetworkName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateRemoteNetworkDestroy,
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
	})
}

func TestAccTwingateRemoteNetworkUpdateWithTheSameName(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Network Update With The Same Name", func(t *testing.T) {
		const terraformResourceName = "test004"
		theResource := acctests.TerraformRemoteNetwork(terraformResourceName)
		name := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateRemoteNetworkDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceRemoteNetwork(terraformResourceName, name),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, name),
						sdk.TestCheckResourceAttr(theResource, locationAttr, model.LocationOther),
					),
				},
				{
					Config: createRemoteNetworkWithLocation(terraformResourceName, name, model.LocationAWS),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, name),
						sdk.TestCheckResourceAttr(theResource, locationAttr, model.LocationAWS),
					),
				},
			},
		})
	})
}
