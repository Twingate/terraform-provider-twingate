package twingate

import (
	"fmt"
	"os"
	"testing"
	"time"

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

	client := NewClient(
		os.Getenv("TWINGATE_URL"),
		os.Getenv("TWINGATE_API_TOKEN"),
		os.Getenv("TWINGATE_NETWORK"),
		30*time.Second,
		2,
		"sweeper")

	return client, nil
}
