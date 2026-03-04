//go:build windows

package diag

import "fmt"

// CheckFirewall is a stub for Windows.
func CheckFirewall() *FirewallStatus {
	return &FirewallStatus{
		Platform: "windows",
		Err:      fmt.Errorf("not supported"),
		ErrMsg:   "not supported",
	}
}
