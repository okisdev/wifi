//go:build windows

package diag

// CheckDFS is a stub on Windows.
func CheckDFS(channel int) *DFSStatus {
	return &DFSStatus{
		Channel: channel,
		IsDFS:   IsDFSChannel(channel),
		State:   "unknown",
	}
}
