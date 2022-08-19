package acctests

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var requiredEnvironmentVariables = []string{
	"TWINGATE_API_TOKEN",
	"TWINGATE_NETWORK",
	"TWINGATE_URL",
}

var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProvider = Provider("test")

	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"twingate": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	t.Run("Test Twingate Resource : Provider", func(t *testing.T) {
		if err := testAccProvider.InternalValidate(); err != nil {
			t.Fatalf("err: %s", err)
		}
	})
}

func testAccPreCheck(t *testing.T) {
	t.Run("Test Twingate Resource : AccPreCheck", func(t *testing.T) {
		for _, requiredEnvironmentVariable := range requiredEnvironmentVariables {
			if value := os.Getenv(requiredEnvironmentVariable); value == "" {
				t.Fatalf("%s must be set before running acceptance tests.", requiredEnvironmentVariable)
			}
		}
	})
}
