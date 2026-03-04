//go:build darwin
package diag

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetConnectionUptime(ifaceName string) time.Duration {
	// Try system_profiler for AirPort info
	out, err := exec.Command("system_profiler", "SPAirPortDataType").Output()
	if err != nil {
		return 0
	}
	// Look for "Last Associated" or similar timestamp - this is unreliable
	// Fallback: use ifconfig uptime
	_ = out

	// Try to get from ipconfig getpacket - contains lease info
	out2, err := exec.Command("ipconfig", "getpacket", ifaceName).Output()
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(out2), "\n") {
		if strings.Contains(line, "lease_time") {
			// Extract lease time in seconds
			parts := strings.Fields(line)
			for _, p := range parts {
				if sec, err := strconv.Atoi(p); err == nil && sec > 0 {
					return time.Duration(sec) * time.Second
				}
			}
		}
	}
	return 0
}
