//go:build linux

package diag

import (
	"os"
	"strings"
)

// GetDNSServers returns configured DNS servers by parsing /etc/resolv.conf on Linux.
func GetDNSServers() *DNSServerInfo {
	info := &DNSServerInfo{Source: "resolv.conf"}

	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		info.Err = err
		info.ErrMsg = err.Error()
		return info
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				info.Servers = append(info.Servers, fields[1])
			}
		}
	}

	return info
}
