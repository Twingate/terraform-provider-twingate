package acctests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

const WaitDuration = 500 * time.Millisecond

func WaitTestFunc() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// Sleep 500 ms
		time.Sleep(WaitDuration)

		return nil
	}
}

func ComposeTestCheckFunc(fs ...resource.TestCheckFunc) resource.TestCheckFunc { //nolint:varnamelen
	return func(state *terraform.State) error {
		if err := WaitTestFunc()(state); err != nil {
			return fmt.Errorf("WaitTestFunc error: %w", err)
		}

		for i, f := range fs {
			if err := f(state); err != nil {
				return fmt.Errorf("check %d/%d error: %w", i+1, len(fs), err)
			}
		}

		return nil
	}
}
