package resource

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func terraformResourceGateway(terraformResourceName, remoteNetworkResourceName, address, x509ResourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway" "%s" {
	  remote_network_id = twingate_remote_network.%s.id
	  address           = "%s"
	  x509_ca_id        = twingate_x509_certificate_authority.%s.id
	}
	`, terraformResourceName, remoteNetworkResourceName, address, x509ResourceName)
}

func terraformResourceGatewayWithSSH(terraformResourceName, remoteNetworkResourceName, address, x509ResourceName, sshResourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway" "%s" {
	  remote_network_id = twingate_remote_network.%s.id
	  address           = "%s"
	  x509_ca_id        = twingate_x509_certificate_authority.%s.id
	  ssh_ca_id         = twingate_ssh_certificate_authority.%s.id
	}
	`, terraformResourceName, remoteNetworkResourceName, address, x509ResourceName, sshResourceName)
}

func gatewayPrerequisites(remoteNetworkName, remoteNetworkTFName, x509TFName, certPEM string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_x509_certificate_authority" "%s" {
	  name        = "%s"
	  certificate = <<-EOT
%s
	EOT
	}
	`, remoteNetworkTFName, remoteNetworkName, x509TFName, test.RandomName(), certPEM)
}

func gatewayPrerequisitesWithSSH(remoteNetworkName, remoteNetworkTFName, x509TFName, certPEM, sshTFName string, publicKey string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_x509_certificate_authority" "%s" {
	  name        = "%s"
	  certificate = <<-EOT
%s
	EOT
	}
	resource "twingate_ssh_certificate_authority" "%s" {
	  name       = "%s"
	  public_key = "%s"
	}
	`, remoteNetworkTFName, remoteNetworkName, x509TFName, test.RandomName(), certPEM, sshTFName, test.RandomName(), publicKey)
}

func TestAccTwingateGatewayCreate(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	gatewayTFName := test.TerraformRandName("test_gw")
	theResource := acctests.TerraformGateway(gatewayTFName)
	address := "10.0.0.1:8080"
	certPEM := acctests.GenerateCACertPEM(t)

	prereqs := gatewayPrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					sdk.TestCheckResourceAttr(theResource, attr.Address, address),
					sdk.TestCheckResourceAttrSet(theResource, attr.RemoteNetworkID),
					sdk.TestCheckResourceAttrSet(theResource, attr.X509CAID),
				),
			},
		},
	})
}

func TestAccTwingateGatewayUpdateAddress(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	gatewayTFName := test.TerraformRandName("test_gw")
	theResource := acctests.TerraformGateway(gatewayTFName)
	address1 := "10.0.0.1:9000"
	address2 := "10.0.0.2:9001"
	certPEM := acctests.GenerateCACertPEM(t)
	resourceID := new(string)

	prereqs := gatewayPrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address1, x509TFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Address, address1),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address2, x509TFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Address, address2),
					// Same ID = in-place update
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value != *resourceID {
							return errors.New("resource should not be re-created on address change")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayUpdateX509CA(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName1 := test.TerraformRandName("test_x509_1")
	x509TFName2 := test.TerraformRandName("test_x509_2")
	gatewayTFName := test.TerraformRandName("test_gw")
	theResource := acctests.TerraformGateway(gatewayTFName)
	certPEM1 := acctests.GenerateCACertPEM(t)
	certPEM2 := acctests.GenerateCACertPEM(t)
	address := "10.0.0.1:8000"
	resourceID := new(string)

	prereqs := fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_x509_certificate_authority" "%s" {
	  name        = "%s"
	  certificate = <<-EOT
%s
	EOT
	}
	resource "twingate_x509_certificate_authority" "%s" {
	  name        = "%s"
	  certificate = <<-EOT
%s
	EOT
	}
	`, remoteNetworkTFName, test.RandomName(), x509TFName1, test.RandomName(), certPEM1, x509TFName2, test.RandomName(), certPEM2)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					// Same ID = in-place update
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value != *resourceID {
							return errors.New("resource should not be re-created on x509_ca_id change")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayAddSSHCA(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshTFName := test.TerraformRandName("test_ssh")
	gatewayTFName := test.TerraformRandName("test_gw")
	theResource := acctests.TerraformGateway(gatewayTFName)
	address := "10.0.0.1:9001"
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)

	prereqs := gatewayPrerequisitesWithSSH(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshTFName, publicKey)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, attr.SSHCAID),
				),
			},
			{
				Config: prereqs + terraformResourceGatewayWithSSH(gatewayTFName, remoteNetworkTFName, address, x509TFName, sshTFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrSet(theResource, attr.SSHCAID),
				),
			},
		},
	})
}

func TestAccTwingateGatewayRemoveSSHCA(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshTFName := test.TerraformRandName("test_ssh")
	gatewayTFName := test.TerraformRandName("test_gw")
	theResource := acctests.TerraformGateway(gatewayTFName)
	address := "10.0.0.1:8001"
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)

	prereqs := gatewayPrerequisitesWithSSH(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshTFName, publicKey)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceGatewayWithSSH(gatewayTFName, remoteNetworkTFName, address, x509TFName, sshTFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrSet(theResource, attr.SSHCAID),
				),
			},
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, attr.SSHCAID),
				),
			},
		},
	})
}

func TestAccTwingateGatewayDelete(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	gatewayTFName := test.TerraformRandName("test_gw")
	theResource := acctests.TerraformGateway(gatewayTFName)
	address := "gateway.twingate.com:8000"
	certPEM := acctests.GenerateCACertPEM(t)

	prereqs := gatewayPrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName),
				Destroy: true,
			},
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateGatewayReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	gatewayTFName := test.TerraformRandName("test_gw")
	theResource := acctests.TerraformGateway(gatewayTFName)
	address := "gateway.twingate.xyz:9008"
	certPEM := acctests.GenerateCACertPEM(t)

	prereqs := gatewayPrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateGatewayDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateGateway),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: prereqs + terraformResourceGateway(gatewayTFName, remoteNetworkTFName, address, x509TFName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Address, address),
				),
			},
		},
	})
}
