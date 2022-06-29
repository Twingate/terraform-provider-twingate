//go:build tools
// +build tools

package tools

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/mattn/goveralls"
	_ "github.com/securego/gosec/v2/cmd/gosec"
	_ "gotest.tools/gotestsum"
)
