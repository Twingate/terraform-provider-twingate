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

// resolvedRegionalURLs caches lookups that came back with a non-error response.
// Terraform calls Configure once per CLI operation, so without this every
// operation repeats the redirect round trip. Failed lookups are not cached:
// they fall back to the unresolved URL, and caching that would pin the process
// to the redirect host for the rest of its lifetime.
var resolvedRegionalURLs sync.Map //nolint:gochecknoglobals

// ResolveRegionalURL returns the regional URL without a slash at the end.
func ResolveRegionalURL(ctx context.Context, network, url string, timeout time.Duration, retryMax int, apiToken, agent, version string) string {
	key := regionalURLKey{network: network, url: url}
	originalURL := SafeURL(fmt.Sprintf("https://%s.%s", network, url))

	return resolveRegionalURL(ctx, originalURL, key, timeout, retryMax, apiToken, agent, version)
}

// resolveRegionalURL performs the lookup for an already-built URL. It is split
// out from ResolveRegionalURL so that tests can point it at a local server,
// which avoids needing a locally trusted certificate for the https:// host that
// ResolveRegionalURL constructs.
func resolveRegionalURL(ctx context.Context, originalURL string, key regionalURLKey, timeout time.Duration, retryMax int, apiToken, agent, version string) string {
	if cached, ok := resolvedRegionalURLs.Load(key); ok {
		return cached.(string) //nolint:forcetypeassert
	}

	correlationID, _ := uuid.GenerateUUID()

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

	// A non-retryable error response arrives here with err == nil, and its final
	// host is the unresolved one, so caching it would pin the process to the
	// redirect host. 5xx cannot reach this point: retryablehttp retries those and
	// returns an error once the attempts are exhausted.
	if resp.StatusCode < http.StatusBadRequest {
		resolvedRegionalURLs.Store(key, resolvedURL)
	}

	log.Printf("[TWINGATE_LOG] [INFO] Resolved regional URL: %s -> %s", originalURL, resolvedURL) // #nosec G706

	return resolvedURL
}
