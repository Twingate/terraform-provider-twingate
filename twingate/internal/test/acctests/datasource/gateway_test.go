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

func terraformDatasourceGateway(tfName, remoteNetworkName, x509Name, address, certPEM string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[2]s"
	}

	resource "twingate_x509_certificate_authority" "%[1]s" {
	  name        = "%[3]s"
	  certificate = <<-EOF
%[5]s
	EOF
	}

	resource "twingate_gateway" "%[1]s" {
	  remote_network_id = twingate_remote_network.%[1]s.id
	  address           = "%[4]s"
	  x509_ca_id        = twingate_x509_certificate_authority.%[1]s.id
	}

	data "twingate_gateway" "%[1]s" {
	  id = twingate_gateway.%[1]s.id
	}
	`, tfName, remoteNetworkName, x509Name, address, strings.TrimSpace(certPEM))
}

func terraformDatasourceGatewayWithSSH(tfName, remoteNetworkName, x509Name, sshName, address, certPEM, publicKey string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[2]s"
	}

	resource "twingate_x509_certificate_authority" "%[1]s" {
	  name        = "%[3]s"
	  certificate = <<-EOF
%[6]s
	EOF
	}

	resource "twingate_ssh_certificate_authority" "%[1]s" {
	  name       = "%[4]s"
	  public_key = "%[7]s"
	}

	resource "twingate_gateway" "%[1]s" {
	  remote_network_id = twingate_remote_network.%[1]s.id
	  address           = "%[5]s"
	  x509_ca_id        = twingate_x509_certificate_authority.%[1]s.id
	  ssh_ca_id         = twingate_ssh_certificate_authority.%[1]s.id
	}

	data "twingate_gateway" "%[1]s" {
	  id = twingate_gateway.%[1]s.id
	}
	`, tfName, remoteNetworkName, x509Name, sshName, address, strings.TrimSpace(certPEM), publicKey)
}

func TestAccDatasourceTwingateGateway_basic(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_ds")
	theResource := acctests.TerraformGateway(tfName)
	theDatasource := acctests.DatasourceName(datasource.TwingateGateway, tfName)
	address := "10.0.0.1:8080"
	certPEM := acctests.GenerateCACertPEM(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformDatasourceGateway(tfName, test.RandomName(), test.RandomName(), address, certPEM),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theDatasource, attr.ID),
					sdk.TestCheckResourceAttr(theDatasource, attr.Address, address),
					sdk.TestCheckResourceAttrSet(theDatasource, attr.RemoteNetworkID),
					sdk.TestCheckResourceAttrSet(theDatasource, attr.X509CAID),
					sdk.TestCheckNoResourceAttr(theDatasource, attr.SSHCAID),
					sdk.TestCheckResourceAttrPair(theDatasource, attr.ID, theResource, attr.ID),
					sdk.TestCheckResourceAttrPair(theDatasource, attr.Address, theResource, attr.Address),
					sdk.TestCheckResourceAttrPair(theDatasource, attr.RemoteNetworkID, theResource, attr.RemoteNetworkID),
					sdk.TestCheckResourceAttrPair(theDatasource, attr.X509CAID, theResource, attr.X509CAID),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateGateway_withSSH(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_ds_ssh")
	theResource := acctests.TerraformGateway(tfName)
	theDatasource := acctests.DatasourceName(datasource.TwingateGateway, tfName)
	address := "10.0.0.2:8080"
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformDatasourceGatewayWithSSH(tfName, test.RandomName(), test.RandomName(), test.RandomName(), address, certPEM, publicKey),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theDatasource, attr.ID),
					sdk.TestCheckResourceAttr(theDatasource, attr.Address, address),
					sdk.TestCheckResourceAttrSet(theDatasource, attr.SSHCAID),
					sdk.TestCheckResourceAttrPair(theDatasource, attr.SSHCAID, theResource, attr.SSHCAID),
				),
			},
		},
	})
}

func testDatasourceTwingateGatewayDoesNotExist(id string) string {
	return fmt.Sprintf(`
	data "twingate_gateway" "test" {
	  id = "%s"
	}
	`, id)
}

func TestAccDatasourceTwingateGateway_doesNotExist(t *testing.T) {
	t.Parallel()

	gatewayID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Gateway:%d", acctest.RandInt())))

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      testDatasourceTwingateGatewayDoesNotExist(gatewayID),
				ExpectError: regexp.MustCompile("failed to read twingate_gateway"),
			},
		},
	})
}

func TestAccDatasourceTwingateGateway_invalidID(t *testing.T) {
	t.Parallel()

	gatewayID := acctest.RandString(10)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      testDatasourceTwingateGatewayDoesNotExist(gatewayID),
				ExpectError: regexp.MustCompile("failed to read twingate_gateway"),
			},
		},
	})
}
