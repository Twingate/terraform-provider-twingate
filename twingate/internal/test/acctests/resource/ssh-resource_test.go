package resource

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func sshResourcePrerequisites(remoteNetworkName, remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress string) string {
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
	resource "twingate_gateway" "%s" {
	  remote_network_id = twingate_remote_network.%s.id
	  address           = "%s"
	  x509_ca_id        = twingate_x509_certificate_authority.%s.id
	  ssh_ca_id         = twingate_ssh_certificate_authority.%s.id
	}
	`, remoteNetworkTFName, remoteNetworkName, x509TFName, test.RandomName(), certPEM, sshCATFName, test.RandomName(), publicKey, gatewayTFName, remoteNetworkTFName, gatewayAddress, x509TFName, sshCATFName)
}

func terraformResourceSSHResource(tfName, gatewayTFName, remoteNetworkTFName, name, address string) string {
	return fmt.Sprintf(`
	resource "twingate_ssh_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName)
}

func terraformResourceSSHResourceWithUsername(tfName, gatewayTFName, remoteNetworkTFName, name, address, username string) string {
	return fmt.Sprintf(`
	resource "twingate_ssh_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  username          = "%s"
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, username)
}

func TestAccTwingateSSHResource_InvalidAddress(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	sshResTFName := test.TerraformRandName("test_ssh_res")
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "10.0.0.1:9000"
	gatewayAddress := "10.0.0.1:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateSSHResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
				ExpectError: regexp.MustCompile(`address string must be a valid IP or FQDN`),
			},
		},
	})
}

func TestAccTwingateSSHResourceCreate(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	sshResTFName := test.TerraformRandName("test_ssh_res")
	theResource := acctests.TerraformSSHResource(sshResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "10.0.0.1"
	gatewayAddress := "10.0.0.1:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateSSHResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckResourceAttr(theResource, attr.Address, resourceAddress),
					sdk.TestCheckResourceAttrSet(theResource, attr.GatewayID),
					sdk.TestCheckResourceAttrSet(theResource, attr.RemoteNetworkID),
				),
			},
		},
	})
}

func TestAccTwingateSSHResourceUpdateName(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	sshResTFName := test.TerraformRandName("test_ssh_res")
	theResource := acctests.TerraformSSHResource(sshResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name1 := test.RandomName()
	name2 := test.RandomName()
	resourceAddress := "10.0.0.2"
	gatewayAddress := "10.0.0.2:8080"
	resourceID := new(string)

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateSSHResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, name1, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name1),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, name2, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name2),
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value != *resourceID {
							return errors.New("resource should not be re-created on name change")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateSSHResourceUpdateUsername(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	sshResTFName := test.TerraformRandName("test_ssh_res")
	theResource := acctests.TerraformSSHResource(sshResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	username1 := test.RandomName()
	username2 := test.RandomName()
	resourceAddress := "10.0.0.2"
	gatewayAddress := "10.0.0.2:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateSSHResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceSSHResourceWithUsername(sshResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, username1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Username, username1),
				),
			},
			{
				Config: prereqs + terraformResourceSSHResourceWithUsername(sshResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, username2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Username, username2),
				),
			},
		},
	})
}

func TestAccTwingateSSHResourceUpdateAddress(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	sshResTFName := test.TerraformRandName("test_ssh_res")
	theResource := acctests.TerraformSSHResource(sshResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	address1 := "10.0.0.3"
	address2 := "10.0.0.4"
	gatewayAddress := "10.0.0.3:8080"
	resourceID := new(string)

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateSSHResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, resourceName, address1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Address, address1),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, resourceName, address2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Address, address2),
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

func TestAccTwingateSSHResourceDelete(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	sshResTFName := test.TerraformRandName("test_ssh_res")
	theResource := acctests.TerraformSSHResource(sshResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "10.0.0.5"
	gatewayAddress := "10.0.0.5:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)
	config := prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateSSHResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  config,
				Destroy: true,
			},
			{
				Config: config,
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateSSHResourceReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	sshResTFName := test.TerraformRandName("test_ssh_res")
	theResource := acctests.TerraformSSHResource(sshResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "10.0.0.6"
	gatewayAddress := "10.0.0.6:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)
	config := prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateSSHResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: config,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateSSHResource),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: config,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Address, resourceAddress),
				),
			},
		},
	})
}

func TestAccTwingateSSHResourceImport(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	sshResTFName := test.TerraformRandName("test_ssh_res")
	theResource := acctests.TerraformSSHResource(sshResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "10.0.0.7"
	gatewayAddress := "10.0.0.7:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateSSHResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceSSHResource(sshResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				ResourceName:      theResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
