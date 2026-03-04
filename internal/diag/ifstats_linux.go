//go:build linux
package diag

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetInterfaceStats(ifaceName string) *InterfaceStats {
	stats := &InterfaceStats{}
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		stats.Err = err
		return stats
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, ifaceName+":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		fields := strings.Fields(parts[1])
		// /proc/net/dev columns: rx_bytes rx_packets rx_errs rx_drop ... tx_bytes tx_packets tx_errs tx_drop ...
		if len(fields) < 11 {
			stats.Err = fmt.Errorf("unexpected /proc/net/dev fields")
			return stats
		}
		stats.RxPackets, _ = strconv.ParseInt(fields[1], 10, 64)
		stats.RxErrors, _ = strconv.ParseInt(fields[2], 10, 64)
		stats.TxPackets, _ = strconv.ParseInt(fields[9], 10, 64)
		stats.TxErrors, _ = strconv.ParseInt(fields[10], 10, 64)
		return stats
	}
	stats.Err = fmt.Errorf("interface %s not found in /proc/net/dev", ifaceName)
	return stats
}
