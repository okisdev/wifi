//go:build windows

package diag

// GetDNSServers is a stub on Windows.
func GetDNSServers() *DNSServerInfo {
	return &DNSServerInfo{Source: "unsupported"}
}
