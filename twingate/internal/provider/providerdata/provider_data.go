package providerdata

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client"

type Config struct {
	Network string
	URL     string
}

type ProviderData struct {
	Client *client.Client
	Config Config
}
