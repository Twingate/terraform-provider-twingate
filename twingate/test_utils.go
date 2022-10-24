package twingate

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"time"
)

func WaitTestFunc() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Sleep 500 ms
		time.Sleep(time.Duration(500) * time.Millisecond)
		return nil
	}
}

func ComposeTestCheckFunc(fs ...resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_ = WaitTestFunc()(s)

		for i, f := range fs {
			if err := f(s); err != nil {
				return fmt.Errorf("Check %d/%d error: %s", i+1, len(fs), err)
			}
		}

		return nil
	}
}
