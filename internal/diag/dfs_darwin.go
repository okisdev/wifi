//go:build darwin

package diag

import (
	"os/exec"
	"strings"
)

// CheckDFS checks the DFS status of the given channel on macOS.
func CheckDFS(channel int) *DFSStatus {
	result := &DFSStatus{
		Channel: channel,
		IsDFS:   IsDFSChannel(channel),
	}

	if !result.IsDFS {
		result.State = "N/A"
		return result
	}

	out, err := exec.Command("system_profiler", "SPAirPortDataType").CombinedOutput()
	if err != nil {
		result.State = "unknown"
		return result
	}

	lower := strings.ToLower(string(out))
	switch {
	case strings.Contains(lower, "radar"):
		result.State = "radar detected"
	case strings.Contains(lower, "dfs"):
		result.State = "dfs active"
	default:
		result.State = "unknown"
	}

	return result
}
