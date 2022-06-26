package twingate

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatasourceTwingateConnectors_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Connectors Basic", func(t *testing.T) {

		networkName1 := acctest.RandomWithPrefix(testPrefixName)
		networkName2 := acctest.RandomWithPrefix(testPrefixName)
		connectorName := acctest.RandomWithPrefix(testPrefixName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateConnectors(networkName1, connectorName, networkName2, connectorName),
					Check: resource.ComposeTestCheckFunc(
						testOutputLength("my_connectors", 2),
						testOutputItemField("my_connectors", 0, "name", connectorName),
					),
				},
			},
		})
	})
}

func testOutputLength(name string, expectedLength int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		values, ok := rs.Value.([]interface{})
		if !ok {
			return errors.New("output not a list")
		}

		if len(values) != expectedLength {
			return fmt.Errorf(
				"output length '%d', didn't match %d",
				len(values),
				expectedLength)
		}

		return nil
	}
}

func testOutputItemField(name string, itemPosition int, filedName, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		values, ok := rs.Value.([]interface{})
		if !ok {
			return errors.New("output not a list")
		}

		if len(values) <= itemPosition {
			return errors.New("item position out of the list")
		}

		item := values[itemPosition]

		obj, ok := item.(map[string]interface{})
		if !ok {
			return errors.New("item not an object")
		}

		val, ok := obj[filedName]
		if !ok {
			return errors.New("field not found in the object")
		}

		if val.(string) != value {
			return fmt.Errorf(
				"expected %s, got %s",
				value,
				val.(string))
		}

		return nil
	}
}

func testDatasourceTwingateConnectors(networkName1, connectorName1, networkName2, connectorName2 string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test1" {
	  name = "%s"
	}
	resource "twingate_connector" "test1" {
	  remote_network_id = twingate_remote_network.test1.id
	  name = "%s"
	}

	resource "twingate_remote_network" "test2" {
	  name = "%s"
	}
	resource "twingate_connector" "test2" {
	  remote_network_id = twingate_remote_network.test2.id
	  name = "%s"
	}

	data "twingate_connectors" "all" {
	  depends_on = [twingate_connector.test1, twingate_connector.test2]
	}

	output "my_connectors" {
	  value = [for conn in data.twingate_connectors.all.connectors: conn if can(regex("^tf-acc", conn.name))] 
	}
	`, networkName1, connectorName1, networkName2, connectorName2)
}
