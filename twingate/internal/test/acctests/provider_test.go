package acctests

import (
	"testing"
)

func TestProvider(t *testing.T) {
	t.Run("Test Twingate Resource : Provider", func(t *testing.T) {
		if err := Provider.InternalValidate(); err != nil {
			t.Fatalf("err: %s", err)
		}
	})
}
