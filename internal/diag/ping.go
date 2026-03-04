package diag

import (
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
)

func Ping(target string) *PingResult {
	result := &PingResult{Target: target}

	args := []string{"-c", "5", "-W", "3", target}
	if runtime.GOOS == "windows" {
		args = []string{"-n", "5", "-w", "3000", target}
	}

	out, err := exec.Command("ping", args...).CombinedOutput()
	if err != nil {
		// Still try to parse partial output
		result.Err = err
		result.ErrMsg = err.Error()
	}

	output := string(out)
	parsePingOutput(output, result)
	return result
}

func parsePingOutput(output string, result *PingResult) {
	// Parse packet loss
	lossRe := regexp.MustCompile(`([\d.]+)%\s*(packet\s+)?loss`)
	if m := lossRe.FindStringSubmatch(output); len(m) > 1 {
		result.LossPct, _ = strconv.ParseFloat(m[1], 64)
	}

	// Parse rtt min/avg/max/mdev (Unix) or Minimum/Maximum/Average (Windows)
	// Unix: rtt min/avg/max/mdev = 1.234/5.678/9.012/1.234 ms
	rttRe := regexp.MustCompile(`(?:rtt|round-trip)\s+min/avg/max/(?:mdev|stddev)\s*=\s*([\d.]+)/([\d.]+)/([\d.]+)/([\d.]+)`)
	if m := rttRe.FindStringSubmatch(output); len(m) > 4 {
		result.MinMs, _ = strconv.ParseFloat(m[1], 64)
		result.AvgMs, _ = strconv.ParseFloat(m[2], 64)
		result.MaxMs, _ = strconv.ParseFloat(m[3], 64)
		result.JitterMs, _ = strconv.ParseFloat(m[4], 64)
		result.Err = nil
		result.ErrMsg = ""
		return
	}

	// Windows: Minimum = 1ms, Maximum = 5ms, Average = 3ms
	winRe := regexp.MustCompile(`Minimum\s*=\s*(\d+)ms.*Maximum\s*=\s*(\d+)ms.*Average\s*=\s*(\d+)ms`)
	if m := winRe.FindStringSubmatch(output); len(m) > 3 {
		result.MinMs, _ = strconv.ParseFloat(m[1], 64)
		result.MaxMs, _ = strconv.ParseFloat(m[2], 64)
		result.AvgMs, _ = strconv.ParseFloat(m[3], 64)
		result.JitterMs = math.Abs(result.MaxMs-result.MinMs) / 2
		result.Err = nil
		result.ErrMsg = ""
		return
	}

	// Parse individual ping times as fallback
	timeRe := regexp.MustCompile(`time[=<]([\d.]+)\s*ms`)
	matches := timeRe.FindAllStringSubmatch(output, -1)
	if len(matches) > 0 {
		var times []float64
		for _, m := range matches {
			if t, err := strconv.ParseFloat(m[1], 64); err == nil {
				times = append(times, t)
			}
		}
		if len(times) > 0 {
			var sum, min, max float64
			min = times[0]
			max = times[0]
			for _, t := range times {
				sum += t
				if t < min {
					min = t
				}
				if t > max {
					max = t
				}
			}
			result.AvgMs = sum / float64(len(times))
			result.MinMs = min
			result.MaxMs = max
			if len(times) > 1 {
				var variance float64
				for _, t := range times {
					d := t - result.AvgMs
					variance += d * d
				}
				result.JitterMs = math.Sqrt(variance / float64(len(times)))
			}
			result.Err = nil
			result.ErrMsg = ""
		}
	}

	if result.AvgMs == 0 && result.Err == nil {
		result.Err = fmt.Errorf("could not parse ping output")
		result.ErrMsg = "could not parse ping output"
	}
}
