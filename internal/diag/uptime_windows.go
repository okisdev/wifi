//go:build windows
package diag

import "time"

func GetConnectionUptime(_ string) time.Duration {
	return 0
}
