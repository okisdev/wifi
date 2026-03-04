package platform

import (
	"os"
	"os/exec"
	"runtime"
)

// OS returns the current OS name.
func OS() string {
	return runtime.GOOS
}

// IsRoot returns true if running as root/admin.
func IsRoot() bool {
	switch runtime.GOOS {
	case "windows":
		// Check if running as administrator by trying to read a protected path
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		return err == nil
	default:
		return os.Geteuid() == 0
	}
}

// HasCommand checks if a command is available in PATH.
func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// NeedsSudo returns true if the current platform requires root for WiFi scanning.
func NeedsSudo() bool {
	if runtime.GOOS == "linux" && !IsRoot() && HasCommand("iw") {
		return true
	}
	return false
}
