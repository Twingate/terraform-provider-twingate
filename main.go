package main

import (
	"github.com/Twingate/terraform-provider-twingate/twingate"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

var (
	version = "dev"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderAddr: "registry.terraform.io/Twingate/twingate",
		ProviderFunc: func() *schema.Provider {
			return twingate.Provider(version)
		},
	})
}
