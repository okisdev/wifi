package platform

import (
	"context"
	"os/exec"
	"time"
)

// RunCommand runs a command with a timeout and returns its output.
func RunCommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	return string(out), err
}
