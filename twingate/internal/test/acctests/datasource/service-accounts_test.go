package datasource

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	serviceAccountsLen = attr.Len(attr.ServiceAccounts)
	keyIDsLen          = attr.Len(attr.ServiceAccounts, attr.KeyIDs)
)

func TestAccDatasourceTwingateServicesFilterByName(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Services - Filter By Name", func(t *testing.T) {

		name := test.Prefix("orange")
		const (
			terraformResourceName = "dts_service"
			theDatasource         = "data.twingate_service_accounts.out"
		)

		config := []terraformServiceConfig{
			{
				serviceName:           name,
				terraformResourceName: test.TerraformRandName(terraformResourceName),
			},
			{
				serviceName:           test.Prefix("lemon"),
				terraformResourceName: test.TerraformRandName(terraformResourceName),
			},
		}

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateServiceAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: terraformConfig(
						createServices(config),
						datasourceServices(name, config),
					),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(theDatasource, serviceAccountsLen, "1"),
						resource.TestCheckResourceAttr(theDatasource, keyIDsLen, "1"),
						resource.TestCheckResourceAttr(theDatasource, attr.ID, "service-by-name-"+name),
					),
				},
			},
		})
	})
}

func TestAccDatasourceTwingateServicesAll(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Services - All", func(t *testing.T) {
		prefix := test.Prefix() + acctest.RandString(4)
		const (
			terraformResourceName = "dts_service"
			theDatasource         = "data.twingate_service_accounts.out"
		)

		config := []terraformServiceConfig{
			{
				serviceName:           prefix + "_orange",
				terraformResourceName: test.TerraformRandName(terraformResourceName),
			},
			{
				serviceName:           prefix + "_lemon",
				terraformResourceName: test.TerraformRandName(terraformResourceName),
			},
		}

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateServiceAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: filterDatasourceServices(prefix, config),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(theDatasource, attr.ID, "all-services"),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: filterDatasourceServices(prefix, config),
					Check: acctests.ComposeTestCheckFunc(
						testCheckOutputLength("my_services", 2),
					),
				},
			},
		})
	})
}

func TestAccDatasourceTwingateServicesEmptyResult(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Services - Empty Result", func(t *testing.T) {

		const theDatasource = "data.twingate_service_accounts.out"

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateServiceAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: datasourceServices(test.RandomName(), nil),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(theDatasource, serviceAccountsLen, "0"),
					),
				},
			},
		})
	})
}

type terraformServiceConfig struct {
	terraformResourceName, serviceName string
}

func terraformConfig(resources ...string) string {
	return strings.Join(resources, "\n")
}

func datasourceServices(name string, configs []terraformServiceConfig) string {
	var dependsOn string
	ids := getTerraformServiceKeys(configs)

	if ids != "" {
		dependsOn = fmt.Sprintf("depends_on = [%s]", ids)
	}

	return fmt.Sprintf(`
	data "twingate_service_accounts" "out" {
	  name = "%s"

	  %s
	}
	`, name, dependsOn)
}

func createServices(configs []terraformServiceConfig) string {
	return strings.Join(
		utils.Map[terraformServiceConfig, string](configs, func(cfg terraformServiceConfig) string {
			return createServiceKey(cfg.terraformResourceName, cfg.serviceName)
		}),
		"\n",
	)
}

func getTerraformServiceKeys(configs []terraformServiceConfig) string {
	return strings.Join(
		utils.Map[terraformServiceConfig, string](configs, func(cfg terraformServiceConfig) string {
			return acctests.TerraformServiceKey(cfg.terraformResourceName)
		}),
		", ",
	)
}

func createServiceKey(terraformResourceName, serviceName string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_service_account_key" "%s" {
	  service_account_id = twingate_service_account.%s.id
	}
	`, createServiceAccount(terraformResourceName, serviceName), terraformResourceName, terraformResourceName)
}

func createServiceAccount(terraformResourceName, serviceName string) string {
	return fmt.Sprintf(`
	resource "twingate_service_account" "%s" {
	  name = "%s"
	}
	`, terraformResourceName, serviceName)
}

func filterDatasourceServices(prefix string, configs []terraformServiceConfig) string {
	return fmt.Sprintf(`
	%s

	data "twingate_service_accounts" "out" {

	}

	output "my_services" {
	  	value = [for c in data.twingate_service_accounts.out.service_accounts : c if length(regexall("^%s", c.name)) > 0]
	}
	`, createServices(configs), prefix)
}

func TestAccDatasourceTwingateServicesAllCursors(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Services - All Cursors", func(t *testing.T) {
		acctests.SetPageLimit(1)
		prefix := test.Prefix() + acctest.RandString(4)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateServiceAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: datasourceServicesConfig(prefix),
					Check: acctests.ComposeTestCheckFunc(
						testCheckOutputLength("my_services", 2),
					),
				},
			},
		})
	})
}

func datasourceServicesConfig(prefix string) string {
	return fmt.Sprintf(`
    resource "twingate_service_account" "%s_1" {
      name = "%s-1"
    }
    
    resource "twingate_service_account" "%s_2" {
      name = "%s-2"
    }

    data "twingate_service_accounts" "out" {
    	depends_on = [twingate_service_account.%s_1, twingate_service_account.%s_2]
    }
    
    output "my_services" {
      value = [for c in data.twingate_service_accounts.out.service_accounts : c if length(regexall("^%s", c.name)) > 0]
      depends_on = [data.twingate_service_accounts.out]
    }
`, duplicate(prefix, 7)...)
}

func duplicate(val string, n int) []any {
	result := make([]any, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, val)
	}

	return result
}
