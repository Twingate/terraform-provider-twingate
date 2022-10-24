package twingate

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"time"
)

const WaitDuration = 500 * time.Millisecond

func WaitTestFunc() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Sleep 500 ms
		time.Sleep(WaitDuration)

		return nil
	}
}

func ComposeTestCheckFunc(fs ...resource.TestCheckFunc) resource.TestCheckFunc { // nolint:varnamelen
	return func(s *terraform.State) error {
		_ = WaitTestFunc()(s)

		for i, f := range fs {
			if err := f(s); err != nil {
				return fmt.Errorf("check %d/%d error: %w", i+1, len(fs), err)
			}
		}

		return nil
	}
}
