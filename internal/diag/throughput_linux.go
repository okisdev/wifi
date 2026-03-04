//go:build linux

package diag

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// MeasureThroughput measures interface throughput by sampling /proc/net/dev twice
// with a 3-second interval on Linux.
func MeasureThroughput(ifaceName string) *ThroughputSample {
	result := &ThroughputSample{}

	rx1, tx1, err := readProcNetDev(ifaceName)
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		return result
	}

	time.Sleep(3 * time.Second)

	rx2, tx2, err := readProcNetDev(ifaceName)
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		return result
	}

	result.RxBytesPerSec = float64(rx2-rx1) / 3.0
	result.TxBytesPerSec = float64(tx2-tx1) / 3.0
	return result
}

// readProcNetDev reads /proc/net/dev and extracts rx_bytes and tx_bytes for the interface.
// Format after colon: rx_bytes rx_packets rx_errs rx_drop rx_fifo rx_frame rx_compressed rx_multicast tx_bytes ...
// Indices:              0        1          2       3       4       5        6              7            8
func readProcNetDev(ifaceName string) (rxBytes, txBytes int64, err error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return 0, 0, fmt.Errorf("read /proc/net/dev: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, ifaceName+":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			return 0, 0, fmt.Errorf("unexpected /proc/net/dev format")
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			return 0, 0, fmt.Errorf("unexpected /proc/net/dev fields: %d", len(fields))
		}

		rxBytes, err = strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse rx_bytes: %w", err)
		}

		txBytes, err = strconv.ParseInt(fields[8], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse tx_bytes: %w", err)
		}

		return rxBytes, txBytes, nil
	}

	return 0, 0, fmt.Errorf("interface %s not found in /proc/net/dev", ifaceName)
}
