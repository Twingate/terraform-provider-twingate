//go:build tools
// +build tools

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	_ "github.com/wadey/gocovmerge"
	_ "gotest.tools/gotestsum"
)
