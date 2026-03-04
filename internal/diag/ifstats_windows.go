//go:build windows
package diag

import (
	"os/exec"
	"strconv"
	"strings"
)

func GetInterfaceStats(ifaceName string) *InterfaceStats {
	stats := &InterfaceStats{}
	out, err := exec.Command("netsh", "interface", "ipv4", "show", "subinterfaces").Output()
	if err != nil {
		stats.Err = err
		return stats
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ifaceName) {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				stats.TxErrors, _ = strconv.ParseInt(fields[2], 10, 64)
				stats.RxErrors, _ = strconv.ParseInt(fields[3], 10, 64)
			}
		}
	}
	return stats
}
