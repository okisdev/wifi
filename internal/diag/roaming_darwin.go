//go:build darwin

package diag

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckRoaming checks for 802.11r/k/v roaming support on macOS.
func CheckRoaming(ifaceName string) *RoamingSupport {
	result := &RoamingSupport{}

	out, err := exec.Command("system_profiler", "SPAirPortDataType").CombinedOutput()
	if err != nil {
		result.Err = fmt.Errorf("system_profiler: %w", err)
		result.ErrMsg = result.Err.Error()
		return result
	}

	lower := strings.ToLower(string(out))

	if strings.Contains(lower, "802.11r") {
		result.Has80211r = true
	}
	if strings.Contains(lower, "802.11k") {
		result.Has80211k = true
	}
	if strings.Contains(lower, "802.11v") {
		result.Has80211v = true
	}

	return result
}
