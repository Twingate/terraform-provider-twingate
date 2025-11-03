package datasource

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	resourcesLen     = attr.Len(attr.Resources)
	resourceNamePath = attr.Path(attr.Resources, attr.Name)
)

func TestAccDatasourceTwingateResources_basic(t *testing.T) {
	t.Parallel()

	networkName := test.RandomName()
	resourceName := test.RandomResourceName()
	const theDatasource = "data.twingate_resources.out_drs1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateResources(networkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, resourcesLen, "2"),
					resource.TestCheckResourceAttr(theDatasource, resourceNamePath, resourceName),
				),
			},
		},
	})
}

func testDatasourceTwingateResources(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_drs1" {
	  name = "%s"
	}

	resource "twingate_resource" "test_drs1_1" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test_drs1.id
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "RESTRICTED"
	      ports = ["80-83", "85"]
	    }
	    udp = {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	  }
	}

	resource "twingate_resource" "test_drs1_2" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test_drs1.id
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	    udp = {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	  }
	}

	data "twingate_resources" "out_drs1" {
	  name = "%s"

	  depends_on = [twingate_resource.test_drs1_1, twingate_resource.test_drs1_2]
	}
	`, networkName, resourceName, resourceName, resourceName)
}

func TestAccDatasourceTwingateResources_emptyResult(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: testTwingateResourcesDoesNotExists(resourceName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.twingate_resources.out_drs2", resourcesLen, "0"),
				),
			},
		},
	})
}

func testTwingateResourcesDoesNotExists(name string) string {
	return fmt.Sprintf(`
	data "twingate_resources" "out_drs2" {
	  name = "%s"
	}

	output "my_resources_drs2" {
	  value = data.twingate_resources.out_drs2.resources
	}
	`, name)
}

func TestAccDatasourceTwingateResourcesWithMultipleFilters(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testDatasourceResourcesWithMultipleFilters(test.RandomResourceName()),
				ExpectError: regexp.MustCompile("Only one of name.*"),
			},
		},
	})
}

func testDatasourceResourcesWithMultipleFilters(name string) string {
	return fmt.Sprintf(`
	data "twingate_resources" "with-multiple-filters" {
	  name_regexp = "%[1]s"
	  name_contains = "%[1]s"
	}
	`, name)
}

func TestAccDatasourceTwingateResourcesFilterByPrefix(t *testing.T) {
	t.Parallel()

	prefix := test.Prefix()
	resourceName := test.RandomResourceName()
	networkName := test.RandomName()
	theDatasource := "data.twingate_resources." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateResourcesFilter(resourceName, networkName, prefix+"_test_app", prefix+"_one", prefix+"_on", attr.FilterByPrefix),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, resourcesLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, resourceNamePath, prefix+"_one"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateResourcesFilterBySuffix(t *testing.T) {
	t.Parallel()

	prefix := test.Prefix()
	resourceName := test.RandomResourceName()
	networkName := test.RandomName()
	theDatasource := "data.twingate_resources." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateResourcesFilter(resourceName, networkName, "test_app_"+prefix, "one_"+prefix, "e_"+prefix, attr.FilterBySuffix),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, resourcesLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, resourceNamePath, "one_"+prefix),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateResourcesFilterByContains(t *testing.T) {
	t.Parallel()

	prefix := test.Prefix()
	resourceName := test.RandomResourceName()
	networkName := test.RandomName()
	theDatasource := "data.twingate_resources." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateResourcesFilter(resourceName, networkName, prefix+"_test_app", prefix+"_one", prefix+"_on", attr.FilterByContains),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, resourcesLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, resourceNamePath, prefix+"_one"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateResourcesFilterByRegexp(t *testing.T) {
	t.Parallel()

	prefix := acctest.RandString(6)
	resourceName := test.RandomResourceName()
	networkName := test.RandomName()
	theDatasource := "data.twingate_resources." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateResourcesFilter(resourceName, networkName, prefix+"_test_app", prefix+"_one", prefix+".*app", attr.FilterByRegexp),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, resourcesLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, resourceNamePath, prefix+"_test_app"),
				),
			},
		},
	})
}

func testDatasourceTwingateResourcesFilter(resourceName, networkName, name1, name2, name, filter string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[2]s" {
	  name = "%[2]s"
	}

	resource "twingate_resource" "%[1]s_1" {
	  name = "%[3]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[2]s.id
	}

	resource "twingate_resource" "%[1]s_2" {
	  name = "%[4]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[2]s.id
	}

	data "twingate_resources" "%[1]s" {
	  name%[6]s = "%[5]s"

	  depends_on = [twingate_resource.%[1]s_1, twingate_resource.%[1]s_2]
	}
	`, resourceName, networkName, name1, name2, name, filter)
}

func TestAccDatasourceTwingateResourcesWithoutFilters(t *testing.T) {
	t.Parallel()

	prefix := test.Prefix()
	resourceName := test.RandomResourceName()
	networkName := test.RandomName()
	theDatasource := "data.twingate_resources." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateResourcesAll(resourceName, networkName, prefix+"_test_app"),
				Check: acctests.ComposeTestCheckFunc(
					testCheckResourceAttrNotEqual(theDatasource, resourcesLen, "0"),
				),
			},
		},
	})
}

func testDatasourceTwingateResourcesAll(resourceName, networkName, name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[2]s" {
	  name = "%[2]s"
	}

	resource "twingate_resource" "%[1]s" {
	  name = "%[3]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[2]s.id
	}

	data "twingate_resources" "%[1]s" {
	  depends_on = [twingate_resource.%[1]s]
	}
	`, resourceName, networkName, name)
}

func TestAccDatasourceTwingateResourcesFilterByTags(t *testing.T) {
	t.Parallel()

	prefix := acctest.RandString(6)
	suffix := acctest.RandString(6)
	tag := acctest.RandString(6)
	resourceName := test.RandomResourceName()
	networkName := test.RandomName()
	theDatasource := "data.twingate_resources." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateResourcesTagsFilter(resourceName, networkName, prefix+"_test_app", prefix+"_"+suffix, tag),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, resourcesLen, "1"),
					resource.TestCheckResourceAttr(theDatasource, resourceNamePath, prefix+"_"+suffix),
				),
			},
		},
	})
}

func testDatasourceTwingateResourcesTagsFilter(resourceName, networkName, name1, name2, tag string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[2]s" {
	  name = "%[2]s"
	}

	resource "twingate_resource" "%[1]s_1" {
	  name = "%[3]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[2]s.id
	  tags = {
	    team = "example_team"
	  }
	}

	resource "twingate_resource" "%[1]s_2" {
	  name = "%[4]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[2]s.id
	  tags = {
	    owner = "%[5]s"
	  }
	}

	data "twingate_resources" "%[1]s" {
	  tags = {
	    owner = "%[5]s"
	  }
	  name_suffix = "%[4]s"

	  depends_on = [twingate_resource.%[1]s_1, twingate_resource.%[1]s_2]
	}
	`, resourceName, networkName, name1, name2, tag)
}
