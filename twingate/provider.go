package twingate

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/datasource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	DefaultHTTPTimeout  = "10"
	DefaultHTTPMaxRetry = "10"
	DefaultURL          = "twingate.com"

	// EnvAPIToken env var for Token.
	EnvAPIToken     = "TWINGATE_API_TOKEN" //#nosec
	EnvNetwork      = "TWINGATE_NETWORK"
	EnvURL          = "TWINGATE_URL"
	EnvHTTPTimeout  = "TWINGATE_HTTP_TIMEOUT"
	EnvHTTPMaxRetry = "TWINGATE_HTTP_MAX_RETRY"
)

func Provider(version string) *schema.Provider {
	provider := &schema.Provider{
		Schema:       providerOptions(),
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			datasource.TwingateResource:  datasource.Resource(),
			datasource.TwingateResources: datasource.Resources(),
		},
	}
	provider.ConfigureContextFunc = configure(version, provider)

	return provider
}

func providerOptions() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		attr.APIToken: {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
			Description: fmt.Sprintf("The access key for API operations. You can retrieve this\n"+
				"from the Twingate Admin Console ([documentation](https://docs.twingate.com/docs/api-overview)).\n"+
				"Alternatively, this can be specified using the %s environment variable.", EnvAPIToken),
		},
		attr.Network: {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: false,
			Description: fmt.Sprintf("Your Twingate network ID for API operations.\n"+
				"You can find it in the Admin Console URL, for example:\n"+
				"`autoco.twingate.com`, where `autoco` is your network ID\n"+
				"Alternatively, this can be specified using the %s environment variable.", EnvNetwork),
		},
		attr.URL: {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: false,
			Description: fmt.Sprintf("The default is '%s'\n"+
				"This is optional and shouldn't be changed under normal circumstances.", DefaultURL),
		},
		attr.HTTPTimeout: {
			Type:     schema.TypeInt,
			Optional: true,
			Description: fmt.Sprintf("Specifies a time limit in seconds for the http requests made. The default value is %s seconds.\n"+
				"Alternatively, this can be specified using the %s environment variable", DefaultHTTPTimeout, EnvHTTPTimeout),
		},
		attr.HTTPMaxRetry: {
			Type:     schema.TypeInt,
			Optional: true,
			Description: fmt.Sprintf("Specifies a retry limit for the http requests made. The default value is %s.\n"+
				"Alternatively, this can be specified using the %s environment variable", DefaultHTTPMaxRetry, EnvHTTPMaxRetry),
		},
	}
}

func configure(version string, _ *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		apiToken := os.Getenv(EnvAPIToken)
		network := os.Getenv(EnvNetwork)
		url := withDefault(os.Getenv(EnvURL), DefaultURL)
		httpTimeout := mustGetInt(withDefault(os.Getenv(EnvHTTPTimeout), DefaultHTTPTimeout))
		httpMaxRetry := mustGetInt(withDefault(os.Getenv(EnvHTTPMaxRetry), DefaultHTTPMaxRetry))

		apiToken = withDefault(data.Get(attr.APIToken).(string), apiToken)
		network = withDefault(data.Get(attr.Network).(string), network)
		url = withDefault(data.Get(attr.URL).(string), url)
		httpTimeout = withDefault(data.Get(attr.HTTPTimeout).(int), httpTimeout)
		httpMaxRetry = withDefault(data.Get(attr.HTTPMaxRetry).(int), httpMaxRetry)

		if network != "" {
			return client.NewClient(url,
					apiToken,
					network,
					time.Duration(httpTimeout)*time.Second,
					httpMaxRetry,
					version),
				nil
		}

		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Twingate client",
				Detail:   "Unable to create anonymous Twingate client, network has to be provided",
			},
		}
	}
}

func withDefault[T comparable](val, defaultVal T) T {
	var zeroValue T
	if val == zeroValue {
		return defaultVal
	}

	return val
}

func mustGetInt(str string) int {
	if val, err := strconv.Atoi(str); err == nil {
		return val
	}

	return 0
}
