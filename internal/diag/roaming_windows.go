//go:build windows

package diag

// CheckRoaming is a stub on Windows.
func CheckRoaming(ifaceName string) *RoamingSupport {
	return &RoamingSupport{}
}
