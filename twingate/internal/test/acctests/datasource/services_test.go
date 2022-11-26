package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const servicesLen = "services.#"

func TestAccDatasourceTwingateServices_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Services Basic", func(t *testing.T) {

		//networkName := test.RandomName()
		//resourceName := test.RandomResourceName()
		const theDatasource = "data.twingate_services.out"

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			//CheckDestroy:      acctests.CheckTwingateResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateServices("hello"),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(theDatasource, servicesLen, "1"),
						//resource.TestCheckResourceAttr(theDatasource, firstResourceName, resourceName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateServices(name string) string {
	return fmt.Sprintf(`
	data "twingate_services" "out" {
	  name = "%s"
	}
	`, name)
}
