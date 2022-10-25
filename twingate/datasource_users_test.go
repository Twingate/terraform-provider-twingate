package twingate

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatasourceTwingateUsers_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Users Basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateUsers(),
					Check: ComposeTestCheckFunc(
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
