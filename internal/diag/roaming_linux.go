//go:build linux

package diag

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckRoaming checks for 802.11r/k/v roaming support on Linux.
func CheckRoaming(ifaceName string) *RoamingSupport {
	result := &RoamingSupport{}

	out, err := exec.Command("iw", "dev", ifaceName, "scan", "dump").CombinedOutput()
	if err != nil {
		result.Err = fmt.Errorf("iw scan dump: %w", err)
		result.ErrMsg = result.Err.Error()
		return result
	}

	output := string(out)

	// FT (Fast Transition) = 802.11r
	if strings.Contains(output, "FT") {
		result.Has80211r = true
	}

	// RRM (Radio Resource Management) = 802.11k
	if strings.Contains(output, "RRM") {
		result.Has80211k = true
	}

	// BSS Transition = 802.11v
	if strings.Contains(output, "BSS Transition") {
		result.Has80211v = true
	}

	return result
}
