//go:build darwin

package diag

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// MeasureThroughput measures interface throughput by sampling netstat twice
// with a 3-second interval on macOS.
func MeasureThroughput(ifaceName string) *ThroughputSample {
	result := &ThroughputSample{}

	rx1, tx1, err := readNetstatBytes(ifaceName)
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		return result
	}

	time.Sleep(3 * time.Second)

	rx2, tx2, err := readNetstatBytes(ifaceName)
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		return result
	}

	result.RxBytesPerSec = float64(rx2-rx1) / 3.0
	result.TxBytesPerSec = float64(tx2-tx1) / 3.0
	return result
}

// readNetstatBytes runs netstat -I <iface> -b and extracts Ibytes and Obytes.
// Header: Name Mtu Network Address Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll
// Indices:  0    1     2       3      4     5      6      7     8     9     10
func readNetstatBytes(ifaceName string) (rxBytes, txBytes int64, err error) {
	out, err := exec.Command("netstat", "-I", ifaceName, "-b").Output()
	if err != nil {
		return 0, 0, fmt.Errorf("netstat: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("unexpected netstat output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 10 {
		return 0, 0, fmt.Errorf("unexpected netstat fields: %d", len(fields))
	}

	rxBytes, err = strconv.ParseInt(fields[6], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse Ibytes: %w", err)
	}

	txBytes, err = strconv.ParseInt(fields[9], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse Obytes: %w", err)
	}

	return rxBytes, txBytes, nil
}
