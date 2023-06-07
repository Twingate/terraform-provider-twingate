package main

import (
	"context"
	"log"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

const registry = "registry.terraform.io/Twingate/twingate"

var (
	version = "dev"
)

func main() {
	err := providerserver.Serve(context.Background(), twingate.New(version), providerserver.ServeOpts{
		Address: registry,
	})
	if err != nil {
		log.Fatal(err)
	}
}
