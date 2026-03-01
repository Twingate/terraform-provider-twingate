package datasource

import (
	"crypto/ed25519"
	"crypto/rand"
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
	"golang.org/x/crypto/ssh"
)

func generateSSHDatasourcePublicKey(t *testing.T) string {
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

func terraformDatasourceSSHCertificateAuthority(terraformResourceName, name, publicKey string) string {
	return fmt.Sprintf(`
	resource "twingate_ssh_certificate_authority" "%[1]s" {
	  name       = "%[2]s"
	  public_key = "%[3]s"
	}

	data "twingate_ssh_certificate_authority" "%[1]s" {
	  id = twingate_ssh_certificate_authority.%[1]s.id
	}

	output "ca_name" {
	  value = data.twingate_ssh_certificate_authority.%[1]s.name
	}
	`, terraformResourceName, name, publicKey)
}

func TestAccDatasourceTwingateSSHCertificateAuthority_basic(t *testing.T) {
	t.Parallel()

	terraformResourceName := test.TerraformRandName("test_ssh_ds")
	theResource := acctests.TerraformSSHCertificateAuthority(terraformResourceName)
	theDatasource := acctests.DatasourceName(datasource.TwingateSSHCertificateAuthority, terraformResourceName)
	name := test.RandomName()
	publicKey := generateSSHDatasourcePublicKey(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateSSHCertificateAuthorityDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformDatasourceSSHCertificateAuthority(terraformResourceName, name, publicKey),
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

func testDatasourceTwingateSSHCertificateAuthorityDoesNotExist(id string) string {
	return fmt.Sprintf(`
	data "twingate_ssh_certificate_authority" "test" {
	  id = "%s"
	}

	output "ca_name" {
	  value = data.twingate_ssh_certificate_authority.test.name
	}
	`, id)
}

func TestAccDatasourceTwingateSSHCertificateAuthority_doesNotExist(t *testing.T) {
	t.Parallel()

	caID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("SSHCertificateAuthority:%d", acctest.RandInt())))

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      testDatasourceTwingateSSHCertificateAuthorityDoesNotExist(caID),
				ExpectError: regexp.MustCompile("failed to read twingate_ssh_certificate_authority"),
			},
		},
	})
}

func TestAccDatasourceTwingateSSHCertificateAuthority_invalidID(t *testing.T) {
	t.Parallel()

	caID := acctest.RandString(10)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      testDatasourceTwingateSSHCertificateAuthorityDoesNotExist(caID),
				ExpectError: regexp.MustCompile("failed to read twingate_ssh_certificate_authority"),
			},
		},
	})
}
