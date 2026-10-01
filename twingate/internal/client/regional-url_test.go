package client

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveRegionalURLReturnsCachedValue(t *testing.T) {
	const (
		network = "cached-network"
		url     = "example.test"
	)

	key := regionalURLKey{network: network, url: url}
	resolvedRegionalURLs.Store(key, "https://cached-network.us1.example.test")

	t.Cleanup(func() { resolvedRegionalURLs.Delete(key) })

	// A cancelled context guarantees any HTTP attempt would fail, so returning
	// the stored value proves the cache short-circuits before the request.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resolved := ResolveRegionalURL(ctx, network, url, time.Second, 0, "token", "TF", "test")

	assert.Equal(t, "https://cached-network.us1.example.test", resolved)
}

func TestResolveRegionalURLDoesNotCacheFailures(t *testing.T) {
	const (
		network = "unresolvable-network"
		url     = "invalid.test"
	)

	key := regionalURLKey{network: network, url: url}
	t.Cleanup(func() { resolvedRegionalURLs.Delete(key) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resolved := ResolveRegionalURL(ctx, network, url, time.Second, 0, "token", "TF", "test")

	assert.Equal(t, "https://unresolvable-network.invalid.test", resolved)

	_, cached := resolvedRegionalURLs.Load(key)
	assert.False(t, cached, "a failed resolve must not be cached")
}

func TestResolveRegionalURLCachesSuccessfulLookup(t *testing.T) {
	var redirectHits, regionalHits atomic.Int64

	regional := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		regionalHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer regional.Close()

	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectHits.Add(1)
		http.Redirect(w, r, regional.URL+r.URL.Path, http.StatusPermanentRedirect)
	}))
	defer redirect.Close()

	key := regionalURLKey{network: "cache-success", url: "example.test"}
	t.Cleanup(func() { resolvedRegionalURLs.Delete(key) })

	first := resolveRegionalURL(context.Background(), redirect.URL, key, 10*time.Second, 0, "token", "TF", "test")

	// Only the host is taken from the response; the scheme is always https
	// because that is the only scheme the provider talks to in production.
	wantURL := "https://" + strings.TrimPrefix(regional.URL, "http://")

	assert.Equal(t, wantURL, first)
	assert.Equal(t, int64(1), redirectHits.Load())
	assert.Equal(t, int64(1), regionalHits.Load())

	second := resolveRegionalURL(context.Background(), redirect.URL, key, 10*time.Second, 0, "token", "TF", "test")

	assert.Equal(t, first, second)
	assert.Equal(t, int64(1), redirectHits.Load(), "second lookup must come from the cache")
	assert.Equal(t, int64(1), regionalHits.Load(), "second lookup must come from the cache")
}

func TestResolveRegionalURLDoesNotCacheErrorResponse(t *testing.T) {
	var hits atomic.Int64

	// A 4xx is not retried, so it reaches the cache check with a nil error
	// and a final host that was never redirected.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	key := regionalURLKey{network: "cache-error", url: "example.test"}
	t.Cleanup(func() { resolvedRegionalURLs.Delete(key) })

	var logs bytes.Buffer

	prev := log.Writer()
	log.SetOutput(&logs)

	t.Cleanup(func() { log.SetOutput(prev) })

	resolveRegionalURL(context.Background(), server.URL, key, 10*time.Second, 0, "token", "TF", "test")

	_, cached := resolvedRegionalURLs.Load(key)
	assert.False(t, cached, "an error response must not be cached")
	assert.NotContains(t, logs.String(), "Resolved regional URL",
		"an error response must not be logged as a successful resolution")
	assert.Contains(t, logs.String(), "returned status 404")

	resolveRegionalURL(context.Background(), server.URL, key, 10*time.Second, 0, "token", "TF", "test")

	assert.Equal(t, int64(2), hits.Load(), "an uncached lookup must be retried on the next call")
}
