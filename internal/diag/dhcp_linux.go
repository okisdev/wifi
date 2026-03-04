//go:build linux

package diag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetDHCPInfo retrieves DHCP lease information on Linux from lease files.
func GetDHCPInfo(ifaceName string) *DHCPInfo {
	result := &DHCPInfo{}

	// Try dhclient lease files first, then NetworkManager lease files
	patterns := []string{
		"/var/lib/dhcp/dhclient.*.leases",
		"/var/lib/NetworkManager/*.lease",
	}

	var leaseContent string
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil || len(files) == 0 {
			continue
		}
		// Read the last file (most recent)
		data, err := os.ReadFile(files[len(files)-1])
		if err != nil {
			continue
		}
		leaseContent = string(data)
		break
	}

	if leaseContent == "" {
		result.Err = fmt.Errorf("no DHCP lease files found")
		result.ErrMsg = "no DHCP lease files found"
		return result
	}

	for _, line := range strings.Split(leaseContent, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSuffix(line, ";")

		// dhcp-server-identifier 192.168.1.1
		if strings.Contains(line, "dhcp-server-identifier") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				result.ServerIP = parts[len(parts)-1]
			}
		}

		// dhcp-lease-time 86400
		if strings.Contains(line, "dhcp-lease-time") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				result.LeaseTime = parts[len(parts)-1]
			}
		}

		// expire <weekday> <date> <time> (fallback)
		if strings.HasPrefix(line, "expire") && result.LeaseTime == "" {
			if idx := strings.Index(line, " "); idx != -1 {
				result.LeaseTime = strings.TrimSpace(line[idx+1:])
			}
		}
	}

	return result
}
