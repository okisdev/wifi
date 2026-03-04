//go:build linux

package diag

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckDFS checks the DFS status of the given channel on Linux.
func CheckDFS(channel int) *DFSStatus {
	result := &DFSStatus{
		Channel: channel,
		IsDFS:   IsDFSChannel(channel),
	}

	if !result.IsDFS {
		result.State = "N/A"
		return result
	}

	out, err := exec.Command("iw", "reg", "get").CombinedOutput()
	if err != nil {
		result.Err = fmt.Errorf("iw reg get: %w", err)
		result.ErrMsg = result.Err.Error()
		result.State = "unknown"
		return result
	}

	output := strings.ToLower(string(out))
	switch {
	case strings.Contains(output, "dfs-usable"):
		result.State = "usable"
	case strings.Contains(output, "dfs-available"):
		result.State = "available"
	case strings.Contains(output, "dfs-radar"):
		result.State = "radar detected"
	case strings.Contains(output, "no-ir"):
		result.State = "no-ir"
	default:
		result.State = "unknown"
	}

	return result
}
