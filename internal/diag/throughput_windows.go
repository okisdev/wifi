//go:build windows

package diag

import "fmt"

// MeasureThroughput is a stub on Windows.
func MeasureThroughput(ifaceName string) *ThroughputSample {
	return &ThroughputSample{
		Err:    fmt.Errorf("not supported on windows"),
		ErrMsg: "not supported on windows",
	}
}
