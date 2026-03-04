//go:build darwin

package diag

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetDHCPInfo retrieves DHCP lease information on macOS using ipconfig.
func GetDHCPInfo(ifaceName string) *DHCPInfo {
	result := &DHCPInfo{}

	out, err := exec.Command("ipconfig", "getpacket", ifaceName).CombinedOutput()
	if err != nil {
		result.Err = fmt.Errorf("ipconfig getpacket: %w", err)
		result.ErrMsg = result.Err.Error()
		return result
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)

		// server_identifier (ip): 192.168.1.1
		if strings.HasPrefix(line, "server_identifier") {
			if idx := strings.LastIndex(line, ":"); idx != -1 {
				result.ServerIP = strings.TrimSpace(line[idx+1:])
			}
		}

		// lease_time (uint32): 86400
		if strings.HasPrefix(line, "lease_time") {
			if idx := strings.LastIndex(line, ":"); idx != -1 {
				result.LeaseTime = strings.TrimSpace(line[idx+1:])
			}
		}
	}

	return result
}
