package resource

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"golang.org/x/crypto/ssh"
)

func generateSSHPublicKey(t *testing.T) string {
	t.Helper()

	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ED25519 key: %v", err)
	}

	sshPubKey, err := ssh.NewPublicKey(privKey.Public())
	if err != nil {
		t.Fatalf("failed to create SSH public key: %v", err)
	}

	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPubKey)))
}

func terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey string) string {
	return fmt.Sprintf(`
	resource "twingate_ssh_certificate_authority" "%s" {
	  name       = "%s"
	  public_key = "%s"
	}
	`, terraformResourceName, name, publicKey)
}

func TestAccTwingateSSHCertificateAuthorityCreate(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_ssh")
	theResource := acctests.TerraformSSHCertificateAuthority(terraformResourceName)
	publicKey := generateSSHPublicKey(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
					sdk.TestCheckResourceAttr(theResource, attr.PublicKey, publicKey),
					sdk.TestCheckResourceAttrSet(theResource, attr.Fingerprint),
				),
			},
		},
	})
}

func TestAccTwingateSSHCertificateAuthorityNameChange(t *testing.T) {
	t.Parallel()

	name1 := test.RandomName()
	name2 := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_ssh")
	theResource := acctests.TerraformSSHCertificateAuthority(terraformResourceName)
	publicKey := generateSSHPublicKey(t)
	resourceID := new(string)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name1, publicKey),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name1),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name2, publicKey),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name2),
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value == *resourceID {
							return errors.New("resource was not re-created after name change")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateSSHCertificateAuthorityPublicKeyChange(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_ssh")
	theResource := acctests.TerraformSSHCertificateAuthority(terraformResourceName)
	publicKey1 := generateSSHPublicKey(t)
	publicKey2 := generateSSHPublicKey(t)
	resourceID := new(string)

	if publicKey1 == publicKey2 {
		t.Skip("Skipping test as public keys are identical")
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value == *resourceID {
							return errors.New("resource was not re-created after public key change")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateSSHCertificateAuthorityNoChanges(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_ssh")
	theResource := acctests.TerraformSSHCertificateAuthority(terraformResourceName)
	publicKey := generateSSHPublicKey(t)
	resourceID := new(string)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value != *resourceID {
							return errors.New("resource should not be re-created when nothing changes")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateSSHCertificateAuthorityDelete(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_ssh")
	theResource := acctests.TerraformSSHCertificateAuthority(terraformResourceName)
	publicKey := generateSSHPublicKey(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey),
				Destroy: true,
			},
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateSSHCertificateAuthorityReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_ssh")
	theResource := acctests.TerraformSSHCertificateAuthority(terraformResourceName)
	publicKey := generateSSHPublicKey(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateSSHCertificateAuthority),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceSSHCertificateAuthority(terraformResourceName, name, publicKey),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
				),
			},
		},
	})
}

func TestAccTwingateSSHCertificateAuthorityMissingRequiredPublicKeyField(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_ssh")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      terraformResourceSSHCertificateAuthorityWithoutPublicKey(terraformResourceName, test.RandomName()),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
		},
	})
}

func terraformResourceSSHCertificateAuthorityWithoutPublicKey(terraformResourceName, name string) string {
	return fmt.Sprintf(`
	resource "twingate_ssh_certificate_authority" "%s" {
	  name = "%s"
	}
	`, terraformResourceName, name)
}

func TestAccTwingateSSHCertificateAuthorityMissingRequiredNameField(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_ssh")
	publicKey := generateSSHPublicKey(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      terraformResourceSSHCertificateAuthorityWithoutName(terraformResourceName, publicKey),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
		},
	})
}

func terraformResourceSSHCertificateAuthorityWithoutName(terraformResourceName, publicKey string) string {
	return fmt.Sprintf(`
	resource "twingate_ssh_certificate_authority" "%s" {
	  public_key = "%s"
	}
	`, terraformResourceName, publicKey)
}
