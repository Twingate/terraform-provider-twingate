package twingate

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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
