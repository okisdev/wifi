package diag

import "math"

// Weights for each dimension
const (
	wSignal    = 0.20
	wSpeed     = 0.20
	wLatency   = 0.20
	wStability = 0.15
	wChannel   = 0.10
	wSecurity  = 0.10
	wDNS       = 0.05
)

func ComputeScore(report *HealthReport) (int, Grade) {
	var total float64

	// Signal: RSSI -30=100, -90=0
	signalScore := clampScore(float64(report.Interface.RSSI+90) / 60.0 * 100)
	total += signalScore * wSignal

	// Speed: log scale, 100 Mbps = 100, 0 = 0
	var speedScore float64
	if report.Speed != nil && report.Speed.Download > 0 {
		speedScore = clampScore(math.Log10(report.Speed.Download+1) / math.Log10(101) * 100)
	}
	total += speedScore * wSpeed

	// Latency: 10ms = 100, 200ms = 0
	var latencyScore float64
	if report.InternetPing != nil && report.InternetPing.AvgMs > 0 {
		latencyScore = clampScore((200 - report.InternetPing.AvgMs) / 190 * 100)
	}
	total += latencyScore * wLatency

	// Stability: jitter + errors
	var stabilityScore float64 = 100
	if report.InternetPing != nil && report.InternetPing.JitterMs > 0 {
		jitterPenalty := clampScore(report.InternetPing.JitterMs / 50 * 100)
		stabilityScore -= jitterPenalty * 0.5
	}
	if report.InternetPing != nil && report.InternetPing.LossPct > 0 {
		stabilityScore -= report.InternetPing.LossPct * 2
	}
	if report.IfStats != nil {
		errRate := float64(report.IfStats.TxErrors + report.IfStats.RxErrors)
		totalPkts := float64(report.IfStats.TxPackets + report.IfStats.RxPackets)
		if totalPkts > 0 && errRate/totalPkts > 0.01 {
			stabilityScore -= 20
		}
	}
	if stabilityScore < 0 {
		stabilityScore = 0
	}
	total += stabilityScore * wStability

	// Channel
	var channelScore float64 = 80 // default if no scan data
	if report.ChannelHealth != nil {
		channelScore = report.ChannelHealth.Score
	}
	total += channelScore * wChannel

	// Security
	total += float64(report.SecurityScore) * wSecurity

	// DNS
	var dnsScore float64 = 80
	if len(report.DNS) > 0 && report.DNS[0].Err == nil {
		ms := report.DNS[0].DurationMs
		dnsScore = clampScore((500 - ms) / 480 * 100)
	}
	total += dnsScore * wDNS

	// Captive portal penalty
	if report.CaptivePortal {
		total *= 0.5
	}

	score := int(math.Round(total))
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score, gradeFromScore(score)
}

func clampScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func gradeFromScore(score int) Grade {
	switch {
	case score >= 90:
		return GradeA
	case score >= 75:
		return GradeB
	case score >= 60:
		return GradeC
	case score >= 40:
		return GradeD
	default:
		return GradeF
	}
}
