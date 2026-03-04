package diag

import (
	"context"
	"net"
	"time"
)

// CheckDNSLeak queries o-o.myaddr.l.google.com TXT to discover the resolver IP
// that Google sees, then compares it against configuredServers to detect leaks.
func CheckDNSLeak(ctx context.Context, configuredServers []string) *DNSLeakResult {
	result := &DNSLeakResult{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	txts, err := r.LookupTXT(ctx, "o-o.myaddr.l.google.com")
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		return result
	}

	for _, txt := range txts {
		ip := net.ParseIP(txt)
		if ip != nil {
			result.ResolverIPs = append(result.ResolverIPs, ip.String())
		}
	}

	configured := make(map[string]bool, len(configuredServers))
	for _, s := range configuredServers {
		configured[s] = true
	}

	for _, rip := range result.ResolverIPs {
		if !configured[rip] {
			result.IsLeaking = true
			break
		}
	}

	return result
}
