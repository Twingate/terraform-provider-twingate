package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/datasource"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func terraformDatasourceX509CertificateAuthority(terraformResourceName, name, cert string) string {
	return fmt.Sprintf(`
	resource "twingate_x509_certificate_authority" "%[1]s" {
	  name        = "%[2]s"
	  certificate = <<-EOF
%[3]s
	EOF
	}

	data "twingate_x509_certificate_authority" "%[1]s" {
	  id = twingate_x509_certificate_authority.%[1]s.id
	}

	output "ca_name" {
	  value = data.twingate_x509_certificate_authority.%[1]s.name
	}
	`, terraformResourceName, name, strings.TrimSpace(cert))
}

func TestAccDatasourceTwingateX509CertificateAuthority_basic(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_x509_ds")
	theResource := acctests.TerraformX509CertificateAuthority(terraformResourceName)
	theDatasource := acctests.DatasourceName(datasource.TwingateX509CertificateAuthority, terraformResourceName)
	name := test.RandomName()
	cert := acctests.GenerateCACertPEM(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformDatasourceX509CertificateAuthority(terraformResourceName, name, cert),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckOutput("ca_name", name),
					sdk.TestCheckResourceAttr(theDatasource, attr.Name, name),
					sdk.TestCheckResourceAttrSet(theDatasource, attr.Fingerprint),
					sdk.TestCheckResourceAttrPair(theDatasource, attr.Fingerprint, theResource, attr.Fingerprint),
				),
			},
		},
	})
}

func testDatasourceTwingateX509CertificateAuthorityDoesNotExist(id string) string {
	return fmt.Sprintf(`
	data "twingate_x509_certificate_authority" "test" {
	  id = "%s"
	}

	output "ca_name" {
	  value = data.twingate_x509_certificate_authority.test.name
	}
	`, id)
}

func TestAccDatasourceTwingateX509CertificateAuthority_doesNotExist(t *testing.T) {
	t.Parallel()

	caID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("X509CertificateAuthority:%d", acctest.RandInt())))

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      testDatasourceTwingateX509CertificateAuthorityDoesNotExist(caID),
				ExpectError: regexp.MustCompile("failed to read twingate_x509_certificate_authority"),
			},
		},
	})
}

func TestAccDatasourceTwingateX509CertificateAuthority_invalidID(t *testing.T) {
	t.Parallel()

	caID := acctest.RandString(10)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      testDatasourceTwingateX509CertificateAuthorityDoesNotExist(caID),
				ExpectError: regexp.MustCompile("failed to read twingate_x509_certificate_authority"),
			},
		},
	})
}
