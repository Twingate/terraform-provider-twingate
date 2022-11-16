package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func createServiceAccount(resourceName, serviceAccountName string) string {
	return fmt.Sprintf(`
	resource "twingate_service_account" "%s" {
	  name = "%s"
	}
	`, resourceName, serviceAccountName)
}

func TestAccTwingateServiceAccountCreateUpdate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Create/Update", func(t *testing.T) {
		const resourceName = "test01"
		theResource := acctests.ResourceName(resource.TwingateServiceAccount, resourceName)
		nameBefore := test.RandomName()
		nameAfter := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceAccount(resourceName, nameBefore),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, nameBefore),
					),
				},
				{
					Config: createServiceAccount(resourceName, nameAfter),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, nameAfter),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceAccountDeleteNonExisting(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Delete NonExisting", func(t *testing.T) {
		const resourceName = "test02"
		theResource := acctests.ResourceName(resource.TwingateServiceAccount, resourceName)
		name := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config:  createServiceAccount(resourceName, name),
					Destroy: true,
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceDoesNotExists(theResource),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceAccountReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Create After Deletion", func(t *testing.T) {
		const resourceName = "test03"
		theResource := acctests.ResourceName(resource.TwingateServiceAccount, resourceName)
		name := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceAccount(resourceName, name),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						deleteTwingateResource(theResource, resource.TwingateServiceAccount),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: createServiceAccount(resourceName, name),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
					),
				},
			},
		})
	})
}
