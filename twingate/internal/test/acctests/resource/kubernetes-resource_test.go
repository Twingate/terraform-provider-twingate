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

func terraformResourceKubernetesResource(tfName, gatewayTFName, remoteNetworkTFName, name, address string) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName)
}

func terraformResourceKubernetesResourceWithInClusterField(tfName, gatewayTFName, remoteNetworkTFName, name, address string, inCluster bool) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  in_cluster        = %v
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, inCluster)
}

func TestAccTwingateKubernetesResource_InvalidAddress(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
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
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
				ExpectError: regexp.MustCompile(`address string must be a valid IP or FQDN`),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceCreate(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.1:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
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

func TestAccTwingateKubernetesResourceUpdateName(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name1 := test.RandomName()
	name2 := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.2:8080"
	resourceID := new(string)

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, name1, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name1),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, name2, resourceAddress),
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

func TestAccTwingateKubernetesResourceUpdateInCluster(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.2:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResourceWithInClusterField(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, true),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.InCluster, "true"),
				),
			},
			{
				Config: prereqs + terraformResourceKubernetesResourceWithInClusterField(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, false),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.InCluster, "false"),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceDelete(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.5:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)
	config := prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
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

func TestAccTwingateKubernetesResourceReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.6:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)
	config := prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: config,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateKubernetesResource),
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

func TestAccTwingateKubernetesResourceImport(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.7:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
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
