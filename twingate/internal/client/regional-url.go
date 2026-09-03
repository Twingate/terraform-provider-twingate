package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-uuid"
)

// ResolveRegionalURL returns the regional URL without a slash at the end.
func ResolveRegionalURL(ctx context.Context, network, url string, timeout time.Duration, retryMax int, apiToken, agent, version string) string {
	correlationID, _ := uuid.GenerateUUID()
	originalURL := SafeURL(fmt.Sprintf("https://%s.%s", network, url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, originalURL, nil)
	if err != nil {
		log.Printf("[TWINGATE_LOG] [ERR] Failed to build request to resolve regional URL: %v", err)

		return originalURL
	}

	httpClient := NewCustomRetryableClient(timeout, retryMax, apiToken, agent, version, correlationID)
	resp, err := httpClient.Do(req) // #nosec G704

	defer func() {
		if resp == nil {
			return
		}

		if err := resp.Body.Close(); err != nil {
			log.Printf("[TWINGATE_LOG] [ERR] Failed to close response body: %v", err)
		}
	}()

	if err != nil {
		log.Printf("[TWINGATE_LOG] [ERR] Failed to resolve regional URL: %v", err)

		return originalURL
	}

	resolvedURL := SafeURL("https://" + resp.Request.URL.Host)
	log.Printf("[TWINGATE_LOG] [INFO] Resolved regional URL: %s -> %s", originalURL, resolvedURL) // #nosec G706

	return resolvedURL
}
