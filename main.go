package main

import (
	"context"
	"log"

	"github.com/Twingate/terraform-provider-twingate/twingate"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version = "dev"
)

func main() {
	err := providerserver.Serve(
		context.Background(),
		twingate.New(version),
		providerserver.ServeOpts{
			Address: "registry.terraform.io/Twingate/twingate",
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	//plugin.Serve(&plugin.ServeOpts{
	//	ProviderFunc: func() *schema.Provider {
	//		return twingate.Provider(version)
	//	},
	//})
}
