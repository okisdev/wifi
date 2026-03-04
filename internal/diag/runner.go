package diag

import (
	"context"
	"time"

	"github.com/okisdev/wifi/internal/speed"
	"github.com/okisdev/wifi/internal/wifi"
)

var stepNames = []string{
	"Getting interface info",       // 0
	"Detecting gateway",            // 1
	"Pinging gateway",              // 2
	"Pinging internet",             // 3
	"Checking DNS",                 // 4
	"Checking DNS servers",         // 5
	"Checking DNS leak",            // 6
	"Checking DoH/DoT support",     // 7
	"Reading interface stats",      // 8
	"Checking connection uptime",   // 9
	"Checking MTU",                 // 10
	"Checking IPv6",                // 11
	"Checking captive portal",      // 12
	"Fetching network identity",    // 13
	"Running traceroute",           // 14
	"Getting DHCP info",            // 15
	"Checking roaming support",     // 16
	"Checking DFS status",          // 17
	"Checking MAC randomization",   // 18
	"Measuring throughput",         // 19
	"Running network quality test", // 20
	"Running speed test",           // 21
	"Scanning channels",            // 22
	"Scanning gateway ports",       // 23
	"Checking firewall",            // 24
	"Checking ARP table",           // 25
	"Computing score",              // 26
}

// RunAll runs all diagnostics and returns a HealthReport.
// progressFn is called after each step with the step info.
func RunAll(ctx context.Context, scanner wifi.Scanner, tester speed.Tester, opts RunOptions, progressFn func(DiagStep, *HealthReport)) (*HealthReport, error) {
	report := &HealthReport{
		Timestamp: time.Now(),
	}

	total := len(stepNames)
	progress := func(i int) {
		if progressFn != nil {
			progressFn(DiagStep{Index: i, Total: total, Name: stepNames[i]}, report)
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

	// Step 5: DNS servers
	progress(5)
	report.DNSServers = GetDNSServers()
	if report.DNSServers.Err != nil {
		report.Issues = append(report.Issues, "Could not read DNS servers: "+report.DNSServers.ErrMsg)
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 6: DNS leak
	progress(6)
	var configuredServers []string
	if report.DNSServers != nil {
		configuredServers = report.DNSServers.Servers
	}
	report.DNSLeak = CheckDNSLeak(ctx, configuredServers)
	if report.DNSLeak.Err != nil {
		report.Issues = append(report.Issues, "DNS leak check failed: "+report.DNSLeak.ErrMsg)
	} else if report.DNSLeak.IsLeaking {
		report.Issues = append(report.Issues, "DNS leak detected — queries may be routed through unexpected resolvers")
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 7: DoH/DoT
	progress(7)
	report.DoHDoT = CheckDoHDoT(ctx)

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 8: Interface stats
	progress(8)
	report.IfStats = GetInterfaceStats(info.Name)
	if report.IfStats.Err != nil {
		report.Issues = append(report.Issues, "Could not read interface stats: "+report.IfStats.ErrMsg)
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 9: Uptime
	progress(9)
	report.Uptime = GetConnectionUptime(info.Name)

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 10: MTU
	progress(10)
	report.MTU = GetMTU(info.Name)
	if report.MTU == 0 {
		report.Issues = append(report.Issues, "Could not determine MTU")
	}

	// Step 11: IPv6
	progress(11)
	report.IPv6 = HasIPv6(info.Name)

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 12: Captive portal
	progress(12)
	report.CaptivePortal = CheckCaptivePortal(ctx)
	if report.CaptivePortal {
		report.Issues = append(report.Issues, "Captive portal detected — internet access may be restricted")
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 13: Network identity (skip if captive portal detected or opted out)
	progress(13)
	if !opts.NoIdentity && !report.CaptivePortal {
		report.Identity = CheckIdentity(ctx)
		if report.Identity.Err != nil {
			report.Issues = append(report.Issues, "Could not fetch network identity: "+report.Identity.ErrMsg)
		} else if report.Identity.IsVPN {
			report.Issues = append(report.Issues, "VPN or privacy relay detected — some results may differ from direct connection")
		}
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 14: Traceroute
	progress(14)
	if !opts.NoTraceroute {
		report.Traceroute = RunTraceroute(ctx)
		if report.Traceroute.Err != nil {
			report.Issues = append(report.Issues, "Traceroute failed: "+report.Traceroute.ErrMsg)
		}
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 15: DHCP
	progress(15)
	report.DHCP = GetDHCPInfo(info.Name)
	if report.DHCP.Err != nil {
		// DHCP errors are informational, don't add to issues
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 16: Roaming support
	progress(16)
	report.Roaming = CheckRoaming(info.Name)

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 17: DFS status
	progress(17)
	report.DFS = CheckDFS(info.Channel)

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 18: MAC randomization
	progress(18)
	report.MACRandom = CheckMACRandomization(info.BSSID)

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 19: Throughput measurement (3s sample)
	progress(19)
	report.Throughput = MeasureThroughput(info.Name)
	if report.Throughput.Err != nil {
		// Informational, don't add to issues
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 20: Network quality (macOS only)
	progress(20)
	if !opts.NoNetworkQuality {
		report.NetworkQuality = RunNetworkQuality(ctx)
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 21: Speed test
	progress(21)
	if !opts.NoSpeed {
		result, err := tester.Run(false, false, nil)
		if err != nil {
			report.Issues = append(report.Issues, "Speed test failed: "+err.Error())
		} else {
			report.Speed = result
		}
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 22: Channel scan & analysis
	progress(22)
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

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 23: Gateway port scan
	progress(23)
	if !opts.NoPortScan && gatewayIP != "" {
		report.GatewayPorts = ScanGatewayPorts(ctx, gatewayIP)
		if report.GatewayPorts.Err != nil {
			report.Issues = append(report.Issues, "Port scan failed: "+report.GatewayPorts.ErrMsg)
		} else {
			for _, p := range report.GatewayPorts.OpenPorts {
				if p.Port == 23 || p.Port == 21 || p.Port == 161 {
					report.Issues = append(report.Issues,
						"Insecure port open on gateway: "+p.Service)
				}
			}
		}
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 24: Firewall
	progress(24)
	report.Firewall = CheckFirewall()
	if report.Firewall.Err == nil && !report.Firewall.Enabled {
		report.Issues = append(report.Issues, "Local firewall is disabled")
	}

	if ctx.Err() != nil {
		return report, ctx.Err()
	}

	// Step 25: ARP spoof detection
	progress(25)
	report.ARPSpoof = CheckARPSpoof()
	if report.ARPSpoof.Err == nil && report.ARPSpoof.IsSuspicious {
		report.Issues = append(report.Issues, "Possible ARP spoofing detected — duplicate MAC addresses in ARP table")
	}

	// Security score (base from protocol + new penalties)
	report.SecurityScore = ComputeSecurityScore(info.Security, report)

	if report.SecurityScore < 50 {
		report.Issues = append(report.Issues, "Weak security score")
	}

	// Efficiency
	dlMbps := 0.0
	if report.Speed != nil {
		dlMbps = report.Speed.Download
	}
	report.Efficiency = CalcEfficiency(dlMbps, info.PHYMode, info.Width, info.TxRate)

	// Step 26: Compute score
	progress(26)
	report.Score, report.Grade = ComputeScore(report)

	return report, nil
}
