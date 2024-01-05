package main

import (
	"context"
	"flag"
	"log"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version = "dev"
)

const registry = "registry.terraform.io/Twingate/twingate"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	err := providerserver.Serve(context.Background(), twingate.New(version),
		providerserver.ServeOpts{
			Debug:           debug,
			Address:         registry,
			ProtocolVersion: 6,
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
