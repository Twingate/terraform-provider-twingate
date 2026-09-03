package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-uuid"
)

type regionalURLKey struct {
	network string
	url     string
}

// resolvedRegionalURLs memoizes successful lookups. Terraform calls Configure
// once per CLI operation, so without this every operation repeats the redirect
// round trip. Only successful resolves are stored: the failure path falls back
// to the unresolved URL, and caching that would pin the process to the redirect
// host after a single transient error.
var resolvedRegionalURLs sync.Map //nolint:gochecknoglobals

// ResolveRegionalURL returns the regional URL without a slash at the end.
func ResolveRegionalURL(ctx context.Context, network, url string, timeout time.Duration, retryMax int, apiToken, agent, version string) string {
	memoKey := regionalURLKey{network: network, url: url}
	if cached, ok := resolvedRegionalURLs.Load(memoKey); ok {
		return cached.(string) //nolint:forcetypeassert
	}

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
		httpClient.CloseIdleConnections()

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
	resolvedRegionalURLs.Store(memoKey, resolvedURL)
	log.Printf("[TWINGATE_LOG] [INFO] Resolved regional URL: %s -> %s", originalURL, resolvedURL) // #nosec G706

	return resolvedURL
}
