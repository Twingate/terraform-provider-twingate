package twingate

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedClient returns a common TwingateClient setup needed for the sweeper
func sharedClient(tenant string) (*Client, error) {
	if os.Getenv("TWINGATE_API_TOKEN") == "" {
		return nil, fmt.Errorf("must provide environment variable TWINGATE_API_TOKEN")
	}

	if os.Getenv("TWINGATE_NETWORK") == "" {
		return nil, fmt.Errorf("must provide environment variable TWINGATE_NETWORK")
	}

	if os.Getenv("TWINGATE_URL") == "" {
		return nil, fmt.Errorf("must provide environment variable TWINGATE_URL")
	}

	client := NewClient(os.Getenv("TWINGATE_NETWORK"), os.Getenv("TWINGATE_API_TOKEN"), os.Getenv("TWINGATE_URL"))

	return client, nil
}
