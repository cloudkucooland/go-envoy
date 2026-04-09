package envoy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/brutella/dnssd"
)

func Discover(ctx context.Context) (string, error) {
	var discovered string

	searchCtx, cancelSearch := context.WithCancel(ctx)
	defer cancelSearch()

	found := func(e dnssd.BrowseEntry) {
		for _, ipa := range e.IPs {
			if ip := ipa.To4(); ip != nil {
				discovered = ip.String()
				slog.DebugContext(ctx, "envoy discovered", "host", discovered, "name", e.Name)
				cancelSearch()
				return
			}
		}
	}

	reject := func(e dnssd.BrowseEntry) {
		slog.InfoContext(ctx, "dnssd entry rejected", "entry", e.Name)
	}

	err := dnssd.LookupType(searchCtx, "_enphase-envoy._tcp.local.", found, reject)

	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return "", fmt.Errorf("dnssd lookup failed: %w", err)
	}

	if discovered == "" {
		return "", fmt.Errorf("no envoy discovered on local network")
	}

	return discovered, nil
}
