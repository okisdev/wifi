package diag

import (
	"context"
	"net"
	"time"
)

var dnsTargets = []struct {
	name     string
	resolver string
}{
	{"google.com", "system"},
}

func CheckDNS(ctx context.Context) []*DNSResult {
	var results []*DNSResult
	for _, target := range dnsTargets {
		r := &DNSResult{Resolver: target.name}
		resolver := &net.Resolver{}

		start := time.Now()
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, err := resolver.LookupHost(ctx2, target.name)
		cancel()
		r.DurationMs = float64(time.Since(start).Microseconds()) / 1000.0

		if err != nil {
			r.Err = err
			r.ErrMsg = err.Error()
		}
		results = append(results, r)
	}
	return results
}
