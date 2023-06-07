package test

import (
	"fmt"
	"os"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
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
	if os.Getenv(twingate.EnvAPIToken) == "" {
		return nil, fmt.Errorf("must provide environment variable %s", twingate.EnvAPIToken)
	}

	if os.Getenv(twingate.EnvNetwork) == "" {
		return nil, fmt.Errorf("must provide environment variable %s", twingate.EnvNetwork)
	}

	if os.Getenv(twingate.EnvURL) == "" {
		return nil, fmt.Errorf("must provide environment variable %s", twingate.EnvURL)
	}

	return client.NewClient(
			os.Getenv(twingate.EnvURL),
			os.Getenv(twingate.EnvAPIToken),
			os.Getenv(twingate.EnvNetwork),
			getHTTPTimeout(twingate.EnvHTTPTimeout, testTimeoutDuration),
			testHTTPRetry,
			"test"),
		nil
}
