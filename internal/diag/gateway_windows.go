//go:build windows
package diag

import (
	"fmt"
	"os/exec"
	"strings"
)

func DetectGateway() (string, error) {
	out, err := exec.Command("cmd", "/c", "route", "print", "0.0.0.0").Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "0.0.0.0" {
			return fields[2], nil
		}
	}
	return "", fmt.Errorf("gateway not found")
}
