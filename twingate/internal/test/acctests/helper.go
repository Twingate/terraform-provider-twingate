package acctests

import (
	"os"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var requiredEnvironmentVariables = []string{
	twingate.EnvAPIToken,
	twingate.EnvNetwork,
	twingate.EnvURL,
}

var Provider *schema.Provider
var ProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	Provider = twingate.Provider("test")

	ProviderFactories = map[string]func() (*schema.Provider, error){
		"twingate": func() (*schema.Provider, error) {
			return Provider, nil
		},
	}
}

func PreCheck(t *testing.T) {
	t.Run("Test Twingate Resource : AccPreCheck", func(t *testing.T) {
		for _, requiredEnvironmentVariable := range requiredEnvironmentVariables {
			if value := os.Getenv(requiredEnvironmentVariable); value == "" {
				t.Fatalf("%s must be set before running acceptance tests.", requiredEnvironmentVariable)
			}
		}
	})
}
