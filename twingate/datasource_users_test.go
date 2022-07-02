package twingate

import (
	"errors"
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
					Check: resource.ComposeTestCheckFunc(
						testCheckOutputNonEmptyArray("all_users"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateUsers() string {
	return fmt.Sprintf(`
	data "twingate_users" "all" {}

	output "all_users" {
	  value = data.twingate_users.all.users
	}
	`)
}

func testCheckOutputNonEmptyArray(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		values, ok := rs.Value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array, got %T", rs.Value)
		}

		if len(values) == 0 {
			return errors.New("got empty array")
		}

		return nil
	}
}
