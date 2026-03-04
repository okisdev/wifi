//go:build linux

package diag

import (
	"os/exec"
	"strings"
)

// CheckFirewall checks the local firewall status on Linux.
// It first tries ufw, then falls back to iptables.
func CheckFirewall() *FirewallStatus {
	result := &FirewallStatus{}

	// Try ufw first.
	out, err := exec.Command("ufw", "status").Output()
	if err == nil {
		result.Platform = "ufw"
		if strings.Contains(strings.ToLower(string(out)), "active") {
			result.Enabled = true
		}
		return result
	}

	// Fall back to iptables.
	out, err = exec.Command("iptables", "-L", "-n").Output()
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		return result
	}

	result.Platform = "iptables"
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 3 {
		result.Enabled = true
	}

	return result
}
