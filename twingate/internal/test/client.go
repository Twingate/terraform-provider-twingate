package test

import (
	"os"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/client"
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
	return client.NewClient(
			os.Getenv(twingate.EnvURL),
			os.Getenv(twingate.EnvAPIToken),
			os.Getenv(twingate.EnvNetwork),
			getHTTPTimeout(twingate.EnvHTTPTimeout, testTimeoutDuration),
			testHTTPRetry,
			"test"),
		nil
}
