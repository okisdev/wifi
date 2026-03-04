//go:build windows

package diag

import "fmt"

// GetDHCPInfo is a stub on Windows.
func GetDHCPInfo(ifaceName string) *DHCPInfo {
	return &DHCPInfo{
		Err:    fmt.Errorf("not supported on windows"),
		ErrMsg: "not supported on windows",
	}
}
