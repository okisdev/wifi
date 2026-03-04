package diag

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// CheckDoHDoT tests whether DNS-over-HTTPS and DNS-over-TLS are reachable.
func CheckDoHDoT(ctx context.Context) *DoHDoTResult {
	result := &DoHDoTResult{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// DoH: query Google's DNS-over-HTTPS endpoint
	result.DoHSupported = checkDoH(ctx)

	// DoT: attempt TLS connection to Cloudflare's DNS-over-TLS endpoint
	result.DoTSupported = checkDoT(ctx)

	return result
}

func checkDoH(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://dns.google/resolve?name=example.com&type=A", nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	return strings.Contains(string(body), "Answer")
}

func checkDoT(ctx context.Context) bool {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", "1.1.1.1:853", &tls.Config{
		ServerName: "cloudflare-dns.com",
	})
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
