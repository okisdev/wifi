package diag

import (
	"context"
	"time"

	"github.com/okisdev/wifi/internal/speed"
	"github.com/okisdev/wifi/internal/wifi"
)

var stepNames = []string{
	"Getting interface info",
	"Detecting gateway",
	"Pinging gateway",
	"Pinging internet",
	"Checking DNS",
	"Reading interface stats",
	"Checking connection uptime",
	"Checking MTU",
	"Checking IPv6",
	"Checking captive portal",
	"Running network quality test",
	"Running speed test",
	"Scanning channels",
	"Computing score",
}

// RunAll runs all diagnostics and returns a HealthReport.
// progressFn is called after each step with the step info.
func RunAll(ctx context.Context, scanner wifi.Scanner, tester speed.Tester, opts RunOptions, progressFn func(DiagStep)) (*HealthReport, error) {
	report := &HealthReport{
		Timestamp: time.Now(),
	}

	total := len(stepNames)
	progress := func(i int) {
		if progressFn != nil {
			progressFn(DiagStep{Index: i, Total: total, Name: stepNames[i]})
		}
	}

	// Step 0: Interface info
	progress(0)
	info, err := scanner.InterfaceInfo()
	if err != nil {
		return nil, err
	}
	report.Interface = info
	report.SNR = info.SNR

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 1: Gateway
	progress(1)
	gatewayIP, err := DetectGateway()
	if err != nil {
		report.Issues = append(report.Issues, "Could not detect gateway: "+err.Error())
	}
	report.GatewayIP = gatewayIP

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 2: Ping gateway
	progress(2)
	if gatewayIP != "" {
		report.GatewayPing = Ping(gatewayIP)
		if report.GatewayPing.Err != nil {
			report.Issues = append(report.Issues, "Gateway ping failed: "+report.GatewayPing.ErrMsg)
		}
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 3: Ping internet
	progress(3)
	report.InternetPing = Ping("8.8.8.8")
	if report.InternetPing.Err != nil {
		report.Issues = append(report.Issues, "Internet ping failed: "+report.InternetPing.ErrMsg)
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 4: DNS
	progress(4)
	report.DNS = CheckDNS(ctx)
	for _, d := range report.DNS {
		if d.Err != nil {
			report.Issues = append(report.Issues, "DNS resolution failed for "+d.Resolver+": "+d.ErrMsg)
		}
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 5: Interface stats
	progress(5)
	report.IfStats = GetInterfaceStats(info.Name)
	if report.IfStats.Err != nil {
		report.Issues = append(report.Issues, "Could not read interface stats: "+report.IfStats.ErrMsg)
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 6: Uptime
	progress(6)
	report.Uptime = GetConnectionUptime(info.Name)

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 7: MTU
	progress(7)
	report.MTU = GetMTU(info.Name)
	if report.MTU == 0 {
		report.Issues = append(report.Issues, "Could not determine MTU")
	}

	// Step 8: IPv6
	progress(8)
	report.IPv6 = HasIPv6(info.Name)

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 9: Captive portal
	progress(9)
	report.CaptivePortal = CheckCaptivePortal(ctx)
	if report.CaptivePortal {
		report.Issues = append(report.Issues, "Captive portal detected — internet access may be restricted")
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 10: Network quality (macOS only)
	progress(10)
	if !opts.NoNetworkQuality {
		report.NetworkQuality = RunNetworkQuality(ctx)
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 11: Speed test
	progress(11)
	if !opts.NoSpeed {
		result, err := tester.Run(false, false)
		if err != nil {
			report.Issues = append(report.Issues, "Speed test failed: "+err.Error())
		} else {
			report.Speed = result
		}
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 12: Channel scan & analysis
	progress(12)
	networks, err := scanner.Scan()
	if err != nil {
		report.Issues = append(report.Issues, "Channel scan failed: "+err.Error())
	} else {
		report.ChannelHealth = AnalyzeChannel(info.Channel, info.Band, networks)
		if report.ChannelHealth.NeighborCount >= 3 {
			report.Issues = append(report.Issues,
				"Channel congestion: multiple networks on your channel")
		}
	}

	// Step 13: Security score
	report.SecurityScore = ScoreSecurity(info.Security)
	if report.SecurityScore < 50 {
		report.Issues = append(report.Issues, "Weak security protocol")
	}

	// Efficiency
	dlMbps := 0.0
	if report.Speed != nil {
		dlMbps = report.Speed.Download
	}
	report.Efficiency = CalcEfficiency(dlMbps, info.PHYMode, info.Width, info.TxRate)

	// Step 13: Compute score
	progress(13)
	report.Score, report.Grade = ComputeScore(report)

	return report, nil
}
