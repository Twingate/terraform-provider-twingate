package datasource

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/datasource"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func terraformDatasourceSyncToS3(terraformResourceName, syncType string) string {
	return fmt.Sprintf(`
	data "twingate_sync_to_s3" "%[1]s" {
	  type = "%[2]s"
	}

	output "oidc_url" {
	  value = data.twingate_sync_to_s3.%[1]s.oidc_url
	}

	output "oidc_prefix" {
	  value = data.twingate_sync_to_s3.%[1]s.oidc_prefix
	}
	`, terraformResourceName, syncType)
}

func TestAccDatasourceTwingateSyncToS3_oidc(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_sync_s3_oidc_ds")
	theDatasource := acctests.DatasourceName(datasource.TwingateSyncToS3, terraformResourceName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: terraformDatasourceSyncToS3(terraformResourceName, model.SyncToS3TypeOIDC),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theDatasource, attr.ID),
					sdk.TestCheckResourceAttr(theDatasource, attr.Type, model.SyncToS3TypeOIDC),
					sdk.TestMatchResourceAttr(theDatasource, attr.OidcURL, regexp.MustCompile(`^https://.+/oidc/v2$`)),
					sdk.TestMatchResourceAttr(theDatasource, attr.OidcPrefix, regexp.MustCompile(`^[^/]+/oidc/v2$`)),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateSyncToS3_iam(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_sync_s3_iam_ds")
	theDatasource := acctests.DatasourceName(datasource.TwingateSyncToS3, terraformResourceName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: terraformDatasourceSyncToS3(terraformResourceName, model.SyncToS3TypeIAM),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theDatasource, attr.ID),
					sdk.TestCheckResourceAttr(theDatasource, attr.Type, model.SyncToS3TypeIAM),
					sdk.TestCheckNoResourceAttr(theDatasource, attr.OidcURL),
					sdk.TestCheckNoResourceAttr(theDatasource, attr.OidcPrefix),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateSyncToS3_invalidType(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_sync_s3_invalid_ds")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      terraformDatasourceSyncToS3(terraformResourceName, "not_a_real_type"),
				ExpectError: regexp.MustCompile(`(?i)must be one of`),
			},
		},
	})
}
