package diag

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// RunTraceroute performs a traceroute to 8.8.8.8 and returns parsed hops.
func RunTraceroute(ctx context.Context) *TracerouteResult {
	target := "8.8.8.8"
	result := &TracerouteResult{Target: target}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "tracert", "-d", "-h", "30", "-w", "2000", target)
	} else {
		cmd = exec.CommandContext(ctx, "traceroute", "-n", "-m", "30", "-w", "2", target)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		// Still try to parse partial output
	}

	hops := parseTracerouteOutput(string(out))
	result.Hops = hops

	if len(hops) == 0 && result.Err == nil {
		result.Err = ctx.Err()
		if result.Err != nil {
			result.ErrMsg = result.Err.Error()
		}
	}

	return result
}

func parseTracerouteOutput(output string) []TracerouteHop {
	var hops []TracerouteHop

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// First field must be a hop number
		hopNum, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		hop := TracerouteHop{Number: hopNum}

		// Check if this is a timeout hop (first token after number is "*")
		if fields[1] == "*" {
			hop.Timeout = true
			hops = append(hops, hop)
			continue
		}

		// Parse IP address (second field)
		hop.Address = fields[1]

		// Parse first RTT (third field), strip " ms" suffix
		if len(fields) >= 3 {
			rttStr := strings.TrimSuffix(fields[2], "ms")
			rttStr = strings.TrimSpace(rttStr)
			if rtt, err := strconv.ParseFloat(rttStr, 64); err == nil {
				hop.RTTMs = rtt
			}
		}

		hops = append(hops, hop)
	}

	return hops
}
