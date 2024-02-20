package datasource

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatasourceTwingateUsers_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Users Basic", func(t *testing.T) {
		acctests.SetPageLimit(1)

		users, err := acctests.GetTestUsers()
		if err != nil && !errors.Is(err, acctests.ErrResourceNotFound) {
			t.Skip("can't run test:", err)
		}

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateUsers(),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_users.all", attr.Len(attr.Users), fmt.Sprintf("%d", len(users))),
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
