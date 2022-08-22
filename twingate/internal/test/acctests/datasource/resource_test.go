package datasource

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatasourceTwingateResource_basic(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Resource Basic", func(t *testing.T) {
		networkName := test.RandomName()
		resourceName := test.RandomResourceName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateResource(networkName, resourceName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_resource.out", "name", resourceName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateResource(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}

	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
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

	data "twingate_resource" "out" {
	  id = twingate_resource.test.id
	}
	`, networkName, resourceName)
}

func testAccCheckTwingateResourceDestroy(s *terraform.State) error {
	client := acctests.Provider.Meta().(*transport.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_resource" {
			continue
		}

		resourceId := rs.Primary.ID

		err := client.DeleteResource(context.Background(), resourceId)
		// expecting error here , since the resource is already gone
		if err == nil {
			return fmt.Errorf("resource with ID %s still present : ", resourceId)
		}
	}

	return nil
}

func TestAccDatasourceTwingateResource_negative(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Resource - does not exists", func(t *testing.T) {
		resourceID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Resource:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
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
	data "twingate_resource" "test" {
	  id = "%s"
	}

	output "my_resource" {
	  value = data.twingate_resource.test.name
	}
	`, id)
}

func TestAccDatasourceTwingateResource_invalidID(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Resource - failed parse resource ID", func(t *testing.T) {
		networkID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
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
