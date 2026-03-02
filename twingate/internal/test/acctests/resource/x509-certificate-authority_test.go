package resource

import (
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
)

func terraformResourceX509CertificateAuthority(terraformResourceName, name, cert string) string {
	return fmt.Sprintf(`
	resource "twingate_x509_certificate_authority" "%s" {
	  name        = "%s"
	  certificate  = <<-EOF
%s
	EOF
	}
	`, terraformResourceName, name, strings.TrimSpace(cert))
}

func TestAccTwingateX509CertificateAuthorityCreate(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_x509")
	theResource := acctests.TerraformX509CertificateAuthority(terraformResourceName)
	cert := acctests.GenerateCACertPEM(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name, cert),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
					sdk.TestCheckResourceAttrSet(theResource, attr.Fingerprint),
					sdk.TestCheckNoResourceAttr(theResource, attr.Certificate),
				),
			},
		},
	})
}

func TestAccTwingateX509CertificateAuthorityNameChange(t *testing.T) {
	t.Parallel()

	name1 := test.RandomName()
	name2 := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_x509")
	theResource := acctests.TerraformX509CertificateAuthority(terraformResourceName)
	cert := acctests.GenerateCACertPEM(t)
	resourceID := new(string)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name1, cert),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name1),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name2, cert),
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

func TestAccTwingateX509CertificateAuthorityCertWithoutChanges(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_x509")
	theResource := acctests.TerraformX509CertificateAuthority(terraformResourceName)
	cert1 := acctests.GenerateCACertPEM(t)
	resourceID := new(string)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name, cert1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name, cert1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value != *resourceID {
							return errors.New("resource should not re-creat when certificate doesn't change")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateX509CertificateAuthorityCertChange(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_x509")
	theResource := acctests.TerraformX509CertificateAuthority(terraformResourceName)
	cert1 := acctests.GenerateCACertPEM(t)
	cert2 := acctests.GenerateCACertPEM(t)
	resourceID := new(string)

	if cert1 == cert2 {
		t.Skip("Skipping test as certificates are identical")
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name, cert1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name, cert2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value == *resourceID {
							return errors.New("resource was not re-created after certificate change")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateX509CertificateAuthorityDelete(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_x509")
	theResource := acctests.TerraformX509CertificateAuthority(terraformResourceName)
	cert := acctests.GenerateCACertPEM(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  terraformResourceX509CertificateAuthority(terraformResourceName, name, cert),
				Destroy: true,
			},
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name, cert),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateX509CertificateAuthorityReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	name := test.RandomName()
	terraformResourceName := test.TerraformRandName("test_x509")
	theResource := acctests.TerraformX509CertificateAuthority(terraformResourceName)
	cert := acctests.GenerateCACertPEM(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name, cert),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateX509CertificateAuthority),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceX509CertificateAuthority(terraformResourceName, name, cert),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
				),
			},
		},
	})
}

func TestAccTwingateX509CertificateAuthorityMissingRequiredCertificateField(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_x509")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      terraformResourceX509CertificateAuthorityWithoutCert(terraformResourceName, test.RandomName()),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
		},
	})
}

func terraformResourceX509CertificateAuthorityWithoutCert(terraformResourceName, name string) string {
	return fmt.Sprintf(`
	resource "twingate_x509_certificate_authority" "%s" {
	  name = "%s"
	}
	`, terraformResourceName, name)
}

func TestAccTwingateX509CertificateAuthorityMissingRequiredNameField(t *testing.T) {
	t.Parallel()

	cert := acctests.GenerateCACertPEM(t)
	terraformResourceName := test.TerraformRandName("test_x509")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      terraformResourceX509CertificateAuthorityWithoutName(terraformResourceName, cert),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
		},
	})
}

func terraformResourceX509CertificateAuthorityWithoutName(terraformResourceName, cert string) string {
	return fmt.Sprintf(`
	resource "twingate_x509_certificate_authority" "%s" {
	  certificate  = <<-EOF
%s
	EOF
	}
	`, terraformResourceName, cert)
}

func TestAccTwingateX509CertificateAuthorityWithInvalidCertificate(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_x509")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateX509CertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      terraformResourceX509CertificateAuthorityWithInvalidCertificate(terraformResourceName, test.RandomName()),
				ExpectError: regexp.MustCompile("Error: Invalid certificate"),
			},
		},
	})
}

func terraformResourceX509CertificateAuthorityWithInvalidCertificate(terraformResourceName, name string) string {
	return fmt.Sprintf(`
	resource "twingate_x509_certificate_authority" "%s" {
	  name  = "%s"
	  certificate  = <<-EOF
-----BEGIN CERTIFICATE-----
	Invalid certificate
-----END CERTIFICATE-----
	EOF
	}
	`, terraformResourceName, name)
}
