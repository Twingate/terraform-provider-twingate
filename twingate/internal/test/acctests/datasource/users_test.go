package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatasourceTwingateUsers_basic(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Users Basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateUsers(),
					Check: resource.ComposeTestCheckFunc(
						testCheckResourceAttrNotEqual("data.twingate_users.all", "users.#", "0"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateUsers() string {
	return `
	data "twingate_users" "all" {}
	`
}

func testCheckResourceAttrNotEqual(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()

		res, ok := ms.Resources[name]
		if !ok || res == nil || res.Primary == nil {
			return fmt.Errorf("resource '%s' not found", name)
		}

		actual, ok := res.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("attribute '%s' not found", key)
		}

		if actual == value {
			return fmt.Errorf("expected not equal value '%s', but got equal", value)
		}

		return nil
	}
}
