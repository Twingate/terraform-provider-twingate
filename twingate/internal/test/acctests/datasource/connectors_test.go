package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatasourceTwingateConnectors_basic(t *testing.T) {
	t.Parallel()
	//t.SkipNow() // fixed in other PR
	t.Run("Test Twingate Datasource : Acc Connectors Basic", func(t *testing.T) {
		networkName1 := test.RandomName()
		networkName2 := test.RandomName()
		connectorName := test.RandomConnectorName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateConnectors(networkName1, connectorName, networkName2, connectorName, test.Prefix()),
					Check: acctests.ComposeTestCheckFunc(
						testCheckOutputLength("my_connectors", 2),
						testCheckOutputAttr("my_connectors", 0, "name", connectorName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateConnectors(networkName1, connectorName1, networkName2, connectorName2, prefix string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_dcs1" {
		name = "%s"
	}
	resource "twingate_connector" "test_dcs1" {
		remote_network_id = twingate_remote_network.test_dcs1.id
		name = "%s"
	}
	resource "twingate_remote_network" "test_dcs2" {
		name = "%s"
	}
	resource "twingate_connector" "test_dcs2" {
		remote_network_id = twingate_remote_network.test_dcs2.id
		name = "%s"
	}
	data "twingate_connectors" "all" {
		depends_on = [twingate_connector.test_dcs1, twingate_connector.test_dcs2]
	}

	output "my_connectors" {
	  	value = [for c in [for conn in data.twingate_connectors.all : conn if can(conn.*.name)][0] : c if length(regexall("%s.*", c.name)) > 0]
	}
		`, networkName1, connectorName1, networkName2, connectorName2, prefix)
}

func TestAccDatasourceTwingateConnectors_emptyResult(t *testing.T) {
	t.Parallel()
	t.SkipNow() // fixed in other PR
	t.Run("Test Twingate Datasource : Acc Connectors - empty result", func(t *testing.T) {
		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnectorsDoesNotExists(test.Prefix()),
					Check: resource.ComposeTestCheckFunc(
						testCheckOutputLength("my_connectors_dcs2", 0),
					),
				},
			},
		})
	})
}

func testTwingateConnectorsDoesNotExists(prefix string) string {
	return fmt.Sprintf(`
		data "twingate_connectors" "all_dcs2" {}
		output "my_connectors_dcs2" {
			value = [for c in [for conn in data.twingate_connectors.all_dcs2 : conn if can(conn.*.name)][0] : c if length(regexall("%s.*", c.name)) > 0]
		}
	`, prefix)
}

func testCheckOutputLength(name string, length int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()

		res, ok := ms.Outputs[name]
		if !ok || res == nil || res.Value == nil {
			return fmt.Errorf("output '%s' not found", name)
		}

		actual, ok := res.Value.([]interface{})
		if !ok {
			return fmt.Errorf("output '%s' is not a list", name)
		}

		if len(actual) != length {
			return fmt.Errorf("expected length %d, got %d", length, len(actual))
		}

		return nil
	}
}

func testCheckOutputAttr(name string, index int, attr string, expected interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()

		res, ok := ms.Outputs[name]
		if !ok || res == nil || res.Value == nil {
			return fmt.Errorf("output '%s' not found", name)
		}

		list, ok := res.Value.([]interface{})
		if !ok {
			return fmt.Errorf("output '%s' is not a list", name)
		}

		if index >= len(list) {
			return fmt.Errorf("index out of bounds, actual length %d", len(list))
		}

		item := list[index]
		obj, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map, actual is %T", item)
		}

		actual, ok := obj[attr]
		if !ok {
			return fmt.Errorf("attribute '%s' not found", attr)
		}

		if cmp.Equal(actual, expected) {
			return nil
		}

		return fmt.Errorf("not equal: expected '%v', got '%v'", expected, actual)
	}
}
