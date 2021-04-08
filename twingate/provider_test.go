package twingate

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var twingateEnvVars = []string{
	"TWINGATE_TOKEN",
	"TWINGATE_TENANT",
}

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"twingate": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv("TWINGATE_TOKEN"); err == "" {
		t.Fatal("TWINGATE_TOKEN must be set for acceptance tests")
	}
	if err := os.Getenv("TWINGATE_TENANT"); err == "" {
		t.Fatal("TWINGATE_TENANT must be set for acceptance tests")
	}
}
