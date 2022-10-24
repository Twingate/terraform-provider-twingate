package twingate

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"math/rand"
	"time"
)

func WaitTestFunc() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Sleep between 500ms-1s
		n := 500 + rand.Intn(500) // n will be between 0 and 10
		time.Sleep(time.Duration(n) * time.Millisecond)
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
