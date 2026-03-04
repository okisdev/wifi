//go:build darwin

package diag

import (
	"os/exec"
	"regexp"
	"strings"
)

// GetDNSServers returns configured DNS servers by parsing scutil --dns output on macOS.
func GetDNSServers() *DNSServerInfo {
	info := &DNSServerInfo{Source: "scutil"}

	out, err := exec.Command("scutil", "--dns").Output()
	if err != nil {
		info.Err = err
		info.ErrMsg = err.Error()
		return info
	}

	re := regexp.MustCompile(`nameserver\[\d+\]\s*:\s*(.+)`)
	matches := re.FindAllStringSubmatch(string(out), -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		server := strings.TrimSpace(m[1])
		if server != "" && !seen[server] {
			seen[server] = true
			info.Servers = append(info.Servers, server)
		}
	}

	return info
}
