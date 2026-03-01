package datasource

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/datasource"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func generateX509DatasourceTestCACertPEM(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA " + test.RandomName(),
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		t.Fatalf("failed to PEM-encode certificate: %v", err)
	}

	return buf.String()
}

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
	cert := generateX509DatasourceTestCACertPEM(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			// Write-only attributes are only supported in Terraform 1.11 and later.
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: acctests.CheckTwingateX509CertificateAuthorityDestroy,
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
