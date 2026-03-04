//go:build darwin

package diag

import (
	"os/exec"
	"strings"
)

// CheckFirewall checks the macOS Application Firewall status.
func CheckFirewall() *FirewallStatus {
	result := &FirewallStatus{
		Platform: "macOS Application Firewall",
	}

	out, err := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate").Output()
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		return result
	}

	if strings.Contains(strings.ToLower(string(out)), "enabled") {
		result.Enabled = true
	}

	return result
}
