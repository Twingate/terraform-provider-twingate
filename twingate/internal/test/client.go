package test

import (
	"context"
	"os"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
)

const (
	testTimeoutDuration = 30 * time.Second
	testHTTPRetry       = 2
)

func getHTTPTimeout(key string, duration time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		parsedDuration, err := time.ParseDuration(value)
		if err != nil {
			return duration
		}

		return parsedDuration
	}

	return duration
}

func TwingateClient() (*client.Client, error) {
	apiToken := os.Getenv(twingate.EnvAPIToken)
	httpTimeout := getHTTPTimeout(twingate.EnvHTTPTimeout, testTimeoutDuration)

	// Resolve the regional URL once, the same way the provider does in Configure,
	// so every request goes straight to the regional host instead of being
	// redirected on each call.
	regionalURL := client.ResolveRegionalURL(
		context.Background(),
		os.Getenv(twingate.EnvNetwork),
		os.Getenv(twingate.EnvURL),
		httpTimeout,
		testHTTPRetry,
		apiToken,
		client.DefaultAgent,
		"test",
	)

	return client.NewClient(context.Background(),
			regionalURL,
			apiToken,
			httpTimeout,
			testHTTPRetry,
			client.DefaultAgent,
			"test",
			client.CacheOptions{}),
		nil
}
