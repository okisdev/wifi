package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/okisdev/wifi/internal/diag"
	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/speed"
	"github.com/okisdev/wifi/internal/wifi"
	"github.com/spf13/cobra"
)

var (
	noSpeed          bool
	noNetworkQuality bool
	noTraceroute     bool
	noIdentity       bool
	noPortScan       bool
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run a comprehensive WiFi health check",
	RunE:  runHealth,
}

func init() {
	healthCmd.Flags().BoolVar(&noSpeed, "no-speed", false, "Skip speed test")
	healthCmd.Flags().BoolVar(&noNetworkQuality, "no-nq", false, "Skip networkQuality test (macOS)")
	healthCmd.Flags().BoolVar(&noTraceroute, "no-traceroute", false, "Skip traceroute")
	healthCmd.Flags().BoolVar(&noIdentity, "no-identity", false, "Skip public IP/ISP lookup")
	healthCmd.Flags().BoolVar(&noPortScan, "no-portscan", false, "Skip gateway port scan")
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	scanner := wifi.NewScanner()
	tester := speed.NewTester()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	opts := diag.RunOptions{
		NoSpeed:          noSpeed,
		NoNetworkQuality: noNetworkQuality,
		NoTraceroute:     noTraceroute,
		NoIdentity:       noIdentity,
		NoPortScan:       noPortScan,
	}

	report, err := diag.RunAll(ctx, scanner, tester, opts, func(step diag.DiagStep, _ *diag.HealthReport) {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "\r[%d/%d] %s...     ", step.Index+1, step.Total, step.Name)
		}
	})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !jsonOutput {
		fmt.Fprintln(os.Stderr) // clear progress line
	}

	data := buildHealthReportData(report)
	renderer := getRenderer()
	return renderer.RenderHealthReport(os.Stdout, data)
}

func buildHealthReportData(r *diag.HealthReport) *output.HealthReportData {
	d := &output.HealthReportData{
		Grade:         string(r.Grade),
		Score:         r.Score,
		SSID:          r.Interface.SSID,
		Channel:       r.Interface.Channel,
		Band:          string(r.Interface.Band),
		Width:         r.Interface.Width,
		Security:      r.Interface.Security,
		SecurityScore: r.SecurityScore,
		PHYMode:       r.Interface.PHYMode,
		RSSI:          r.Interface.RSSI,
		SignalQuality: output.SignalQuality(r.Interface.RSSI),
		SNR:           r.SNR,
		Efficiency:    r.Efficiency,
		MTU:           r.MTU,
		IPv6:          r.IPv6,
		CaptivePortal: r.CaptivePortal,
		Issues:        r.Issues,
	}

	if r.Speed != nil {
		d.Download = r.Speed.Download
		d.Upload = r.Speed.Upload
		d.SpeedServer = r.Speed.Server
	}

	if r.GatewayPing != nil && r.GatewayPing.Err == nil {
		d.GatewayPing = fmt.Sprintf("%.1f ms avg, %.1f ms jitter, %.1f%% loss",
			r.GatewayPing.AvgMs, r.GatewayPing.JitterMs, r.GatewayPing.LossPct)
	}

	if r.InternetPing != nil && r.InternetPing.Err == nil {
		d.InternetPing = fmt.Sprintf("%.1f ms avg, %.1f ms jitter, %.1f%% loss",
			r.InternetPing.AvgMs, r.InternetPing.JitterMs, r.InternetPing.LossPct)
	}

	if len(r.DNS) > 0 && r.DNS[0].Err == nil {
		d.DNS = fmt.Sprintf("%.0f ms (%s)", r.DNS[0].DurationMs, r.DNS[0].Resolver)
	}

	if r.IfStats != nil && r.IfStats.Err == nil {
		d.TxErrors = r.IfStats.TxErrors
		d.RxErrors = r.IfStats.RxErrors
		d.Collisions = r.IfStats.Collisions
	}

	if r.Uptime > 0 {
		d.Uptime = r.Uptime.String()
	}

	if r.ChannelHealth != nil {
		d.ChannelCurrent = r.ChannelHealth.Channel
		d.ChannelNeighbors = r.ChannelHealth.NeighborCount
		d.Congestion = r.ChannelHealth.CongestionLevel
		d.BestChannel = r.ChannelHealth.BestChannel
	}

	// Network Identity
	if r.Identity != nil && r.Identity.Err == nil {
		d.PublicIP = r.Identity.PublicIP
		d.ISP = r.Identity.ISP
		d.ASN = r.Identity.ASN
		d.IsVPN = r.Identity.IsVPN
		loc := ""
		if r.Identity.City != "" {
			loc = r.Identity.City
		}
		if r.Identity.Country != "" {
			if loc != "" {
				loc += ", "
			}
			loc += r.Identity.Country
		}
		d.GeoLocation = loc
	}

	// DNS Configuration
	if r.DNSServers != nil && r.DNSServers.Err == nil && len(r.DNSServers.Servers) > 0 {
		d.DNSServersList = strings.Join(r.DNSServers.Servers, ", ")
	}
	if r.DNSLeak != nil && r.DNSLeak.Err == nil {
		if r.DNSLeak.IsLeaking {
			d.DNSLeak = "Leak detected"
		} else {
			d.DNSLeak = "No leak"
		}
	}
	if r.DoHDoT != nil && r.DoHDoT.Err == nil {
		d.DoH = r.DoHDoT.DoHSupported
		d.DoT = r.DoHDoT.DoTSupported
	}

	// Traceroute
	if r.Traceroute != nil && r.Traceroute.Err == nil && len(r.Traceroute.Hops) > 0 {
		d.TracerouteHops = len(r.Traceroute.Hops)
		lastHop := r.Traceroute.Hops[len(r.Traceroute.Hops)-1]
		if lastHop.Timeout {
			d.TracerouteFinal = "timeout"
		} else {
			d.TracerouteFinal = fmt.Sprintf("%.1f ms", lastHop.RTTMs)
		}
	}

	// DHCP
	if r.DHCP != nil && r.DHCP.Err == nil {
		d.DHCPServer = r.DHCP.ServerIP
		d.DHCPLease = r.DHCP.LeaseTime
	}

	// Roaming
	if r.Roaming != nil && r.Roaming.Err == nil {
		features := []string{}
		if r.Roaming.Has80211r {
			features = append(features, "802.11r")
		}
		if r.Roaming.Has80211k {
			features = append(features, "802.11k")
		}
		if r.Roaming.Has80211v {
			features = append(features, "802.11v")
		}
		if len(features) > 0 {
			d.Roaming = strings.Join(features, ", ")
		} else {
			d.Roaming = "None detected"
		}
	}

	// DFS
	if r.DFS != nil && r.DFS.Err == nil {
		if r.DFS.IsDFS {
			d.DFS = fmt.Sprintf("Yes (ch %d, %s)", r.DFS.Channel, r.DFS.State)
		} else {
			d.DFS = "No"
		}
	}

	// MAC Randomization
	if r.MACRandom != nil && r.MACRandom.Err == nil {
		if r.MACRandom.Enabled {
			d.MACRandom = "Enabled (locally-administered)"
		} else {
			d.MACRandom = "Disabled (hardware)"
		}
	}

	// Throughput
	if r.Throughput != nil && r.Throughput.Err == nil {
		d.ThroughputTx = r.Throughput.TxBytesPerSec * 8 / 1_000_000
		d.ThroughputRx = r.Throughput.RxBytesPerSec * 8 / 1_000_000
	}

	// Gateway ports
	if r.GatewayPorts != nil && r.GatewayPorts.Err == nil {
		if len(r.GatewayPorts.OpenPorts) == 0 {
			d.GatewayPortsOpen = "None"
		} else {
			ports := []string{}
			for _, p := range r.GatewayPorts.OpenPorts {
				ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Service))
			}
			d.GatewayPortsOpen = strings.Join(ports, ", ")
		}
	}

	// Firewall
	if r.Firewall != nil && r.Firewall.Err == nil {
		if r.Firewall.Enabled {
			d.FirewallStatus = fmt.Sprintf("Enabled (%s)", r.Firewall.Platform)
		} else {
			d.FirewallStatus = fmt.Sprintf("Disabled (%s)", r.Firewall.Platform)
		}
	}

	// ARP spoof
	if r.ARPSpoof != nil && r.ARPSpoof.Err == nil {
		if r.ARPSpoof.IsSuspicious {
			d.ARPStatus = "Suspicious — duplicate MACs"
		} else {
			d.ARPStatus = "OK"
		}
	}

	return d
}
