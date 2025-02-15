package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	connectorsLen     = attr.Len(attr.Connectors)
	connectorNamePath = attr.Path(attr.Connectors, attr.Name)
	connectorIDPath   = attr.Path(attr.Connectors, attr.ID)
)

func TestAccDatasourceTwingateConnectors_basic(t *testing.T) {
	t.Parallel()

	networkName1 := test.RandomName()
	networkName2 := test.RandomName()
	connectorName := test.RandomConnectorName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateConnectors(networkName1, connectorName, networkName2, connectorName, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					testCheckOutputLength("my_connectors", 2),
					testCheckOutputAttrSet("my_connectors", 0, attr.ID),
					testCheckOutputAttr("my_connectors", 0, attr.Name, connectorName),
					testCheckOutputAttr("my_connectors", 0, attr.StatusUpdatesEnabled, true),
					testCheckOutputAttr("my_connectors", 0, attr.State, "DEAD_NO_HEARTBEAT"),
					testCheckOutputAttr("my_connectors", 0, attr.Hostname, ""),
					testCheckOutputAttr("my_connectors", 0, attr.Version, ""),
					testCheckOutputAttr("my_connectors", 0, attr.PublicIP, ""),
					testCheckOutputAttr("my_connectors", 0, attr.PrivateIPs, []any{}),
				),
			},
		},
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
	prefix := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testTwingateConnectorsDoesNotExists(prefix),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputLength("my_connectors_dcs2", 0),
				),
			},
		},
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

func testCheckOutputAttrSet(name string, index int, attr string) resource.TestCheckFunc {
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

		if len(actual.(string)) > 0 {
			return nil
		}

		return fmt.Errorf("got empty: expected not empty value")
	}
}

func TestAccDatasourceTwingateConnectorsFilterByName(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	connectorName := test.RandomConnectorName()
	theDatasource := "data.twingate_connectors." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateConnectorsFilter(resourceName, test.RandomName(), connectorName, "", connectorName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, connectorsLen, "1"),
					resource.TestCheckResourceAttrSet(theDatasource, connectorIDPath),
					resource.TestCheckResourceAttr(theDatasource, connectorNamePath, connectorName),
				),
			},
		},
	})
}

func testDatasourceTwingateConnectorsFilter(resourceName, networkName, connectorName, filter, name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s_network" {
		name = "%[2]s"
	}
	resource "twingate_connector" "%[1]s_connector" {
		remote_network_id = twingate_remote_network.%[1]s_network.id
		name = "%[3]s"
	}
	
	data "twingate_connectors" "%[1]s" {
		name%[4]s = "%[5]s"
		depends_on = [twingate_connector.%[1]s_connector]
	}
	`, resourceName, networkName, connectorName, filter, name)
}

func TestAccDatasourceTwingateConnectorsFilterByPrefix(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	connectorName := test.RandomConnectorName()
	theDatasource := "data.twingate_connectors." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateConnectorsFilter(resourceName, test.RandomName(), connectorName, attr.FilterByPrefix, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, connectorsLen, "1"),
					resource.TestCheckResourceAttrSet(theDatasource, connectorIDPath),
					resource.TestCheckResourceAttr(theDatasource, connectorNamePath, connectorName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateConnectorsFilterBySuffix(t *testing.T) {
	t.Parallel()

	connectorName := test.RandomConnectorName()
	resourceName := test.RandomResourceName()
	theDatasource := "data.twingate_connectors." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateConnectorsFilter(resourceName, test.RandomName(), connectorName, attr.FilterBySuffix, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, connectorsLen, "1"),
					resource.TestCheckResourceAttrSet(theDatasource, connectorIDPath),
					resource.TestCheckResourceAttr(theDatasource, connectorNamePath, connectorName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateConnectorsFilterByContains(t *testing.T) {
	t.Parallel()

	connectorName := test.RandomConnectorName()
	resourceName := test.RandomResourceName()
	theDatasource := "data.twingate_connectors." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateConnectorsFilter(resourceName, test.RandomName(), connectorName, attr.FilterByContains, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, connectorsLen, "1"),
					resource.TestCheckResourceAttrSet(theDatasource, connectorIDPath),
					resource.TestCheckResourceAttr(theDatasource, connectorNamePath, connectorName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateConnectorsFilterByRegexp(t *testing.T) {
	t.Parallel()

	connectorName := test.RandomConnectorName()
	resourceName := test.RandomResourceName()
	theDatasource := "data.twingate_connectors." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateConnectorsFilter(resourceName, test.RandomName(), connectorName, attr.FilterByRegexp, fmt.Sprintf(".*%s.*", connectorName)),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, connectorsLen, "1"),
					resource.TestCheckResourceAttrSet(theDatasource, connectorIDPath),
					resource.TestCheckResourceAttr(theDatasource, connectorNamePath, connectorName),
				),
			},
		},
	})
}
