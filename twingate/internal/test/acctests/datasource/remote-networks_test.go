package datasource

import (
	"fmt"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	remoteNetworksLen     = attr.Len(attr.RemoteNetworks)
	remoteNetworkNamePath = attr.Path(attr.RemoteNetworks, attr.Name)
)

func TestAccDatasourceTwingateRemoteNetworks_read(t *testing.T) {
	t.Parallel()

	prefix := acctest.RandString(10)
	networkName1 := test.RandomName(prefix)
	networkName2 := test.RandomName(prefix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateRemoteNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateRemoteNetworks2(networkName1, networkName2, prefix),
				Check: acctests.ComposeTestCheckFunc(
					testCheckOutputLength("test_networks", 2),
				),
			},
		},
	})
}

func testDatasourceTwingateRemoteNetworks2(networkName1, networkName2, prefix string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_drn1" {
		name = "%s"
	}
	
	resource "twingate_remote_network" "test_drn2" {
		name = "%s"
	}
	
	data "twingate_remote_networks" "all" {
		depends_on = [twingate_remote_network.test_drn1, twingate_remote_network.test_drn2]
	}

	output "test_networks" {
	  	value = [for n in [for net in data.twingate_remote_networks.all : net if can(net.*.name)][6] : n if length(regexall("%s.*", n.name)) > 0]
	}
		`, networkName1, networkName2, prefix)
}

func TestAccDatasourceTwingateRemoteNetworksFilterByName(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	networkName := test.RandomName()
	theDatasource := "data.twingate_remote_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateRemoteNetworksFilter(resourceName, networkName, networkName, ""),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, remoteNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, remoteNetworkNamePath, networkName),
				),
			},
		},
	})
}

func testDatasourceTwingateRemoteNetworksFilter(resourceName, networkName, name, filter string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[2]s"
	}

	data "twingate_remote_networks" "%[1]s" {
	  name%[3]s = "%[4]s"

	  depends_on = [twingate_remote_network.%[1]s]
	}
	`, resourceName, networkName, filter, name)
}

func TestAccDatasourceTwingateRemoteNetworksFilterByPrefix(t *testing.T) {
	t.Parallel()

	prefix := acctest.RandString(5)
	resourceName := test.RandomResourceName()
	networkName := prefix + "_" + test.RandomName()
	theDatasource := "data.twingate_remote_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateRemoteNetworksFilter(resourceName, networkName, prefix, attr.FilterByPrefix),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, remoteNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, remoteNetworkNamePath, networkName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateRemoteNetworksFilterBySuffix(t *testing.T) {
	t.Parallel()

	suffix := acctest.RandString(5)
	resourceName := test.RandomResourceName()
	networkName := test.RandomName() + "_" + suffix
	theDatasource := "data.twingate_remote_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateRemoteNetworksFilter(resourceName, networkName, suffix, attr.FilterBySuffix),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, remoteNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, remoteNetworkNamePath, networkName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateRemoteNetworksFilterByContains(t *testing.T) {
	t.Parallel()

	randString := acctest.RandString(5)
	resourceName := test.RandomResourceName()
	networkName := test.RandomName() + "_" + randString + "_" + acctest.RandString(5)
	theDatasource := "data.twingate_remote_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateRemoteNetworksFilter(resourceName, networkName, randString, attr.FilterByContains),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, remoteNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, remoteNetworkNamePath, networkName),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateRemoteNetworksFilterByRegexp(t *testing.T) {
	t.Parallel()

	randString := acctest.RandString(5)
	resourceName := test.RandomResourceName()
	networkName := test.RandomName() + "_" + randString + "_" + acctest.RandString(5)
	theDatasource := "data.twingate_remote_networks." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateRemoteNetworksFilter(resourceName, networkName, ".*_"+randString+"_.*", attr.FilterByRegexp),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, remoteNetworksLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, remoteNetworkNamePath, networkName),
				),
			},
		},
	})
}
