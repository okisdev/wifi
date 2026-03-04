//go:build darwin
package diag

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func GetInterfaceStats(ifaceName string) *InterfaceStats {
	stats := &InterfaceStats{}
	out, err := exec.Command("netstat", "-I", ifaceName, "-b").Output()
	if err != nil {
		stats.Err = err
		return stats
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		stats.Err = fmt.Errorf("unexpected netstat output")
		return stats
	}
	// Header: Name Mtu Network Address Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll
	fields := strings.Fields(lines[1])
	if len(fields) < 11 {
		stats.Err = fmt.Errorf("unexpected netstat fields: %d", len(fields))
		return stats
	}
	stats.RxPackets, _ = strconv.ParseInt(fields[4], 10, 64)
	stats.RxErrors, _ = strconv.ParseInt(fields[5], 10, 64)
	stats.TxPackets, _ = strconv.ParseInt(fields[7], 10, 64)
	stats.TxErrors, _ = strconv.ParseInt(fields[8], 10, 64)
	stats.Collisions, _ = strconv.ParseInt(fields[10], 10, 64)
	return stats
}
