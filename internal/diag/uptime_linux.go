//go:build linux
package diag

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetConnectionUptime(ifaceName string) time.Duration {
	out, err := exec.Command("iw", "dev", ifaceName, "station", "dump").Output()
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "connected time:") {
			val := strings.TrimPrefix(line, "connected time:")
			val = strings.TrimSpace(strings.TrimSuffix(val, "seconds"))
			if sec, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				return time.Duration(sec) * time.Second
			}
		}
	}
	return 0
}
