//go:build tools
// +build tools

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	_ "github.com/mattn/goveralls"
	_ "gotest.tools/gotestsum"
)
