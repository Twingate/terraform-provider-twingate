package acctests

import (
	"errors"
	"fmt"
	"testing"
	"time"

	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestWaitTestFunc(t *testing.T) {
	start := time.Now()
	err := WaitTestFunc()(nil)
	end := time.Now()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, end.Sub(start), WaitDuration)
}

func TestComposeTestCheckFunc(t *testing.T) {
	badError := errors.New("bad error")

	cases := []struct {
		checkFuncs  []sdk.TestCheckFunc
		expectedErr error
	}{
		{},
		{
			checkFuncs: []sdk.TestCheckFunc{
				func(s *terraform.State) error {
					return nil
				},
				func(s *terraform.State) error {
					return badError
				},
			},
			expectedErr: fmt.Errorf("check 2/2 error: %w", badError),
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			err := ComposeTestCheckFunc(c.checkFuncs...)(nil)

			assert.Equal(t, c.expectedErr, err)
		})
	}

}
