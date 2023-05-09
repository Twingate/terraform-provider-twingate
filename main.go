package main

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version = "dev"
)

func main() {
	providerserver.Serve(context.Background(), twingate.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/Twingate/twingate",
	})
}
