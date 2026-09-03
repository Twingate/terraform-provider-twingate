package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveRegionalURLReturnsMemoizedValue(t *testing.T) {
	const (
		network = "memoized-network"
		url     = "example.test"
	)

	key := regionalURLKey{network: network, url: url}
	resolvedRegionalURLs.Store(key, "https://memoized-network.us1.example.test")

	t.Cleanup(func() { resolvedRegionalURLs.Delete(key) })

	// A cancelled context guarantees any HTTP attempt would fail, so returning
	// the stored value proves the memo short-circuits before the request.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resolved := ResolveRegionalURL(ctx, network, url, time.Second, 0, "token", "TF", "test")

	assert.Equal(t, "https://memoized-network.us1.example.test", resolved)
}

func TestResolveRegionalURLDoesNotMemoizeFailures(t *testing.T) {
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

	_, memoized := resolvedRegionalURLs.Load(key)
	assert.False(t, memoized, "a failed resolve must not be memoized")
}
