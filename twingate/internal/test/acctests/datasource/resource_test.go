package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateResource_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resource Basic", func(t *testing.T) {
		networkName := test.RandomName()
		resourceName := test.RandomResourceName()

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateResource(networkName, resourceName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_resource.out_dr1", attr.Name, resourceName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateResource(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_dr1" {
	  name = "%s"
	}

	resource "twingate_resource" "test_dr1" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test_dr1.id
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "RESTRICTED"
	      ports = ["80-83", "85"]
	    }
	    udp {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	  }
	}

	data "twingate_resource" "out_dr1" {
	  id = twingate_resource.test_dr1.id
	}
	`, networkName, resourceName)
}

func TestAccDatasourceTwingateResource_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resource - does not exists", func(t *testing.T) {
		resourceID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Resource:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateResourceDoesNotExists(resourceID),
					ExpectError: regexp.MustCompile("Error: failed to read resource with id"),
				},
			},
		})
	})
}

func testTwingateResourceDoesNotExists(id string) string {
	return fmt.Sprintf(`
	data "twingate_resource" "test_dr2" {
	  id = "%s"
	}

	output "my_resource_dr2" {
	  value = data.twingate_resource.test_dr2.name
	}
	`, id)
}

func TestAccDatasourceTwingateResource_invalidID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resource - failed parse resource ID", func(t *testing.T) {
		networkID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateResourceDoesNotExists(networkID),
					ExpectError: regexp.MustCompile("Unable to parse global ID"),
				},
			},
		})
	})
}
