package datasource

import (
	"fmt"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

var (
	exitNetworksLen     = attr.Len(attr.ExitNetworks)
	exitNetworkNamePath = attr.Path(attr.ExitNetworks, attr.Name)
)

func TestAccDatasourceTwingateExitNetworks_read(t *testing.T) {
	t.Parallel()

	prefix := acctest.RandString(10)
	networkName1 := test.RandomName(prefix)
	networkName2 := test.RandomName(prefix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateExitNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateExitNetworks2(networkName1, networkName2, prefix),
				Check: acctests.ComposeTestCheckFunc(
					testCheckOutputLength("test_networks", 2),
				),
			},
		},
	})
}

func testDatasourceTwingateExitNetworks2(networkName1, networkName2, prefix string) string {
	return fmt.Sprintf(`
	resource "twingate_exit_network" "test_drn1" {
		name = "%s"
	}

	resource "twingate_exit_network" "test_drn2" {
		name = "%s"
	}

	data "twingate_exit_networks" "all" {
		depends_on = [twingate_exit_network.test_drn1, twingate_exit_network.test_drn2]
	}

	output "test_networks" {
	  	value = [for n in [for net in data.twingate_exit_networks.all : net if can(net.*.name)][0] : n if length(regexall("%s.*", n.name)) > 0]
	}
		`, networkName1, networkName2, prefix)
}

func TestAccDatasourceTwingateExitNetworksFilterByName(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	networkName := test.RandomName()
	theDatasource := "data.twingate_exit_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateExitNetworksFilter(resourceName, networkName, networkName, ""),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, exitNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, exitNetworkNamePath, networkName),
				),
			},
		},
	})
}

func testDatasourceTwingateExitNetworksFilter(resourceName, networkName, name, filter string) string {
	return fmt.Sprintf(`
	resource "twingate_exit_network" "%[1]s" {
	  name = "%[2]s"
	}

	data "twingate_exit_networks" "%[1]s" {
	  name%[3]s = "%[4]s"

	  depends_on = [twingate_exit_network.%[1]s]
	}
	`, resourceName, networkName, filter, name)
}

func TestAccDatasourceTwingateExitNetworksFilterByPrefix(t *testing.T) {
	t.Parallel()

	prefix := acctest.RandString(5)
	resourceName := test.RandomResourceName()
	networkName := prefix + "_" + test.RandomName()
	theDatasource := "data.twingate_exit_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateExitNetworksFilter(resourceName, networkName, prefix, attr.FilterByPrefix),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, exitNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, exitNetworkNamePath, networkName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateExitNetworksFilterBySuffix(t *testing.T) {
	t.Parallel()

	suffix := acctest.RandString(5)
	resourceName := test.RandomResourceName()
	networkName := test.RandomName() + "_" + suffix
	theDatasource := "data.twingate_exit_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateExitNetworksFilter(resourceName, networkName, suffix, attr.FilterBySuffix),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, exitNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, exitNetworkNamePath, networkName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateExitNetworksFilterByContains(t *testing.T) {
	t.Parallel()

	randString := acctest.RandString(5)
	resourceName := test.RandomResourceName()
	networkName := test.RandomName() + "_" + randString + "_" + acctest.RandString(5)
	theDatasource := "data.twingate_exit_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateExitNetworksFilter(resourceName, networkName, randString, attr.FilterByContains),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, exitNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, exitNetworkNamePath, networkName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateExitNetworksFilterByRegexp(t *testing.T) {
	t.Parallel()

	randString := acctest.RandString(5)
	resourceName := test.RandomResourceName()
	networkName := test.RandomName() + "_" + randString + "_" + acctest.RandString(5)
	theDatasource := "data.twingate_exit_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateExitNetworksFilter(resourceName, networkName, ".*_"+randString+"_.*", attr.FilterByRegexp),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, exitNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, exitNetworkNamePath, networkName),
				),
			},
		},
	})
}
