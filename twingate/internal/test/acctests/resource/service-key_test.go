package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	statusAttr = "status"
)

func createServiceKey(terraformResourceName, serviceAccountName string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_service_key" "%s" {
	  service = twingate_service.%s.id
	}
	`, createServiceAccount(terraformResourceName, serviceAccountName), terraformResourceName, terraformResourceName)
}

func createServiceKeyWithName(terraformResourceName, serviceAccountName, serviceKeyName string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_service_key" "%s" {
	  service = twingate_service.%s.id
	  name = "%s"
	}
	`, createServiceAccount(terraformResourceName, serviceAccountName), terraformResourceName, terraformResourceName, serviceKeyName)
}

func TestAccTwingateServiceKeyCreateUpdate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Key Create/Update", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceAccount := acctests.TerraformServiceAccount(terraformResourceName)
		serviceKey := acctests.TerraformServiceKey(terraformResourceName)

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccount),
						sdk.TestCheckResourceAttr(serviceAccount, nameAttr, serviceAccountName),
						acctests.CheckTwingateResourceExists(serviceKey),
						sdk.TestCheckResourceAttr(serviceKey, statusAttr, model.StatusActive),
					),
				},
				{
					Config: createServiceKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccount),
						sdk.TestCheckResourceAttr(serviceAccount, nameAttr, serviceAccountName),
						acctests.CheckTwingateResourceExists(serviceKey),
						sdk.TestCheckResourceAttr(serviceKey, statusAttr, model.StatusActive),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceKeyCreateUpdateWithName(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Key Create/Update With Name", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceAccount := acctests.TerraformServiceAccount(terraformResourceName)
		serviceKey := acctests.TerraformServiceKey(terraformResourceName)
		beforeName := test.RandomName()
		afterName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceKeyWithName(terraformResourceName, serviceAccountName, beforeName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccount),
						sdk.TestCheckResourceAttr(serviceAccount, nameAttr, serviceAccountName),
						acctests.CheckTwingateResourceExists(serviceKey),
						sdk.TestCheckResourceAttr(serviceKey, nameAttr, beforeName),
						sdk.TestCheckResourceAttr(serviceKey, statusAttr, model.StatusActive),
					),
				},
				{
					Config: createServiceKeyWithName(terraformResourceName, serviceAccountName, afterName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccount),
						sdk.TestCheckResourceAttr(serviceAccount, nameAttr, serviceAccountName),
						acctests.CheckTwingateResourceExists(serviceKey),
						sdk.TestCheckResourceAttr(serviceKey, nameAttr, afterName),
						sdk.TestCheckResourceAttr(serviceKey, statusAttr, model.StatusActive),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceKeyReCreateAfterInactive(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Key ReCreate After Inactive", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceKey := acctests.TerraformServiceKey(terraformResourceName)

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceKey),
						acctests.RevokeTwingateServiceKey(serviceKey),
						acctests.WaitTestFunc(),
						acctests.CheckTwingateServiceKeyStatus(serviceKey, model.StatusRevoked),
					),
				},
				{
					Config: createServiceKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceKey),
						acctests.CheckTwingateServiceKeyStatus(serviceKey, model.StatusActive),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceKeyDelete(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Key Delete", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceKey := acctests.TerraformServiceKey(terraformResourceName)

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config:  createServiceKey(terraformResourceName, serviceAccountName),
					Destroy: true,
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceDoesNotExists(serviceKey),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceKeyReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Key ReCreate After Delete", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceKey := acctests.TerraformServiceKey(terraformResourceName)

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceKey),
						acctests.RevokeTwingateServiceKey(serviceKey),
						acctests.DeleteTwingateResource(serviceKey, resource.TwingateServiceAccountKey),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: createServiceKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceKey),
					),
				},
			},
		})
	})
}
