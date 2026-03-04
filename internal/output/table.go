package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TableRenderer renders data as formatted terminal tables.
type TableRenderer struct {
	NoColor bool
}

func NewTableRenderer(noColor bool) *TableRenderer {
	return &TableRenderer{NoColor: noColor}
}

func (r *TableRenderer) RenderNetworks(w io.Writer, nets []NetworkRow) error {
	if len(nets) == 0 {
		fmt.Fprintln(w, "No networks found.")
		return nil
	}

	headers := []string{"SSID", "BSSID", "SIGNAL", "CH", "BAND", "SECURITY", "WIDTH", "QUALITY"}
	widths := []int{24, 17, 8, 4, 6, 10, 6, 10}

	r.printHeader(w, headers, widths)

	for _, n := range nets {
		ssid := n.SSID
		if ssid == "" {
			ssid = "(hidden)"
		}
		if len(ssid) > 23 {
			ssid = ssid[:20] + "..."
		}

		rssiStr := fmt.Sprintf("%d dBm", n.Signal)
		quality := SignalQuality(n.Signal)
		bar := SignalBar(n.Signal, r.NoColor)

		if !r.NoColor {
			color := SignalColor(n.Signal)
			rssiStr = lipgloss.NewStyle().Foreground(color).Render(rssiStr)
			quality = lipgloss.NewStyle().Foreground(color).Render(quality)
		}

		cols := []string{
			ssid,
			n.BSSID,
			rssiStr,
			fmt.Sprintf("%d", n.Channel),
			n.Band,
			n.Security,
			fmt.Sprintf("%dMHz", n.Width),
			bar + " " + quality,
		}
		r.printRow(w, cols, widths)
	}
	return nil
}

func (r *TableRenderer) RenderInterfaceInfo(w io.Writer, info map[string]string) error {
	title := "WiFi Interface"
	if !r.NoColor {
		title = StyleTitle.Render(title)
	}
	fmt.Fprintln(w, title)
	fmt.Fprintln(w, strings.Repeat("─", 45))

	keys := []string{"Interface", "MAC", "SSID", "BSSID", "RSSI", "Noise", "Channel", "Band", "Tx Rate", "Security", "IP", "Status"}
	for _, k := range keys {
		if v, ok := info[k]; ok && v != "" {
			label := fmt.Sprintf("  %-12s", k)
			if !r.NoColor {
				label = StyleDim.Render(label)
			}
			fmt.Fprintf(w, "%s %s\n", label, v)
		}
	}
	return nil
}

func (r *TableRenderer) RenderSpeedResult(w io.Writer, result map[string]string) error {
	title := "Speed Test Results"
	if !r.NoColor {
		title = StyleTitle.Render(title)
	}
	fmt.Fprintln(w, title)
	fmt.Fprintln(w, strings.Repeat("─", 45))

	keys := []string{"Server", "Location", "Latency", "Jitter", "Download", "Upload"}
	for _, k := range keys {
		if v, ok := result[k]; ok && v != "" {
			label := fmt.Sprintf("  %-12s", k)
			if !r.NoColor {
				label = StyleDim.Render(label)
			}
			fmt.Fprintf(w, "%s %s\n", label, v)
		}
	}
	return nil
}

func (r *TableRenderer) RenderSignal(w io.Writer, data SignalData) error {
	ssid := data.SSID
	if ssid == "" {
		ssid = "(unknown)"
	}

	bar := SignalBar(data.RSSI, r.NoColor)
	rssiStr := fmt.Sprintf("%d dBm", data.RSSI)
	quality := SignalQuality(data.RSSI)

	if !r.NoColor {
		color := SignalColor(data.RSSI)
		rssiStr = lipgloss.NewStyle().Foreground(color).Render(rssiStr)
		quality = lipgloss.NewStyle().Foreground(color).Render(quality)
	}

	fmt.Fprintf(w, "  Network    %s\n", ssid)
	fmt.Fprintf(w, "  Signal     %s %s  %s\n", bar, rssiStr, quality)
	fmt.Fprintf(w, "  Range      %d to %d dBm\n", data.Min, data.Max)
	fmt.Fprintf(w, "  Average    %.1f dBm\n", data.Avg)
	fmt.Fprintf(w, "  Jitter     %.1f dB\n", data.Jitter)
	fmt.Fprintf(w, "  Samples    %d\n", data.Samples)
	return nil
}

func (r *TableRenderer) RenderChannelAnalysis(w io.Writer, channels []ChannelInfo) error {
	if len(channels) == 0 {
		fmt.Fprintln(w, "No channel data available.")
		return nil
	}

	title := "Channel Analysis"
	if !r.NoColor {
		title = StyleTitle.Render(title)
	}
	fmt.Fprintln(w, title)

	headers := []string{"CHANNEL", "BAND", "NETWORKS", "AVG RSSI", "RATING"}
	widths := []int{8, 6, 10, 10, 12}
	r.printHeader(w, headers, widths)

	for _, ch := range channels {
		rating := ch.Rating
		if !r.NoColor {
			switch ch.Rating {
			case "Good":
				rating = StyleGood.Render(rating)
			case "Fair":
				rating = StyleWarn.Render(rating)
			case "Congested":
				rating = StyleBad.Render(rating)
			}
		}

		cols := []string{
			fmt.Sprintf("%d", ch.Channel),
			ch.Band,
			fmt.Sprintf("%d", ch.Networks),
			fmt.Sprintf("%d dBm", ch.AvgRSSI),
			rating,
		}
		r.printRow(w, cols, widths)
	}
	return nil
}

func (r *TableRenderer) printHeader(w io.Writer, headers []string, widths []int) {
	line := ""
	for i, h := range headers {
		line += fmt.Sprintf("%-*s", widths[i]+2, h)
	}
	if !r.NoColor {
		line = lipgloss.NewStyle().Bold(true).Faint(true).Render(line)
	}
	fmt.Fprintln(w, line)
}

func (r *TableRenderer) printRow(w io.Writer, cols []string, widths []int) {
	line := ""
	for i, c := range cols {
		if i < len(widths) {
			// For columns with ANSI codes, we can't rely on %-*s
			// Just pad with spaces based on visible width
			line += c + strings.Repeat(" ", max(2, widths[i]+2-visibleLen(c)))
		} else {
			line += c
		}
	}
	fmt.Fprintln(w, line)
}

func (r *TableRenderer) RenderHealthReport(w io.Writer, data *HealthReportData) error {
	// Header
	gradeStr := fmt.Sprintf("Grade: %s (%d)", data.Grade, data.Score)
	title := fmt.Sprintf("Wi-Fi Health Report — %s", gradeStr)
	if !r.NoColor {
		title = StyleTitle.Render(title)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  "+title)
	fmt.Fprintln(w, "  "+strings.Repeat("─", 50))

	// Section: Network Info
	r.printSection(w, "Network Info")
	r.printKV(w, "SSID", data.SSID)
	r.printKV(w, "Channel", fmt.Sprintf("%d (%s, %dMHz)", data.Channel, data.Band, data.Width))
	r.printKV(w, "Security", fmt.Sprintf("%-20s %d/100", data.Security, data.SecurityScore))
	if data.PHYMode != "" {
		r.printKV(w, "PHY Mode", data.PHYMode)
	}

	rssiStr := fmt.Sprintf("%d dBm (%s)  %s", data.RSSI, data.SignalQuality, SignalBar(data.RSSI, r.NoColor))
	if !r.NoColor {
		rssiStr = lipgloss.NewStyle().Foreground(SignalColor(data.RSSI)).Render(
			fmt.Sprintf("%d dBm (%s)", data.RSSI, data.SignalQuality)) + "  " + SignalBar(data.RSSI, r.NoColor)
	}
	r.printKV(w, "Signal", rssiStr)
	if data.SNR > 0 {
		r.printKV(w, "SNR", fmt.Sprintf("%d dB", data.SNR))
	}

	// Section: Network Identity
	if data.PublicIP != "" {
		r.printSection(w, "Network Identity")
		r.printKV(w, "Public IP", data.PublicIP)
		if data.ISP != "" {
			r.printKV(w, "ISP", data.ISP)
		}
		if data.ASN != "" {
			r.printKV(w, "ASN", data.ASN)
		}
		if data.GeoLocation != "" {
			r.printKV(w, "Location", data.GeoLocation)
		}
		if data.IsVPN {
			vpnStr := "Detected"
			if !r.NoColor {
				vpnStr = StyleWarn.Render(vpnStr)
			}
			r.printKV(w, "VPN/Relay", vpnStr)
		}
	}

	// Section: Speed Test
	if data.Download > 0 || data.Upload > 0 {
		r.printSection(w, "Speed Test")
		if data.SpeedServer != "" {
			r.printKV(w, "Server", data.SpeedServer)
		}
		if data.Download > 0 {
			dl := fmt.Sprintf("%.2f Mbps", data.Download)
			if !r.NoColor {
				dl = StyleGood.Render(dl)
			}
			r.printKV(w, "Download", dl)
		}
		if data.Upload > 0 {
			ul := fmt.Sprintf("%.2f Mbps", data.Upload)
			if !r.NoColor {
				ul = lipgloss.NewStyle().Foreground(lipgloss.Color("#00bfff")).Render(ul)
			}
			r.printKV(w, "Upload", ul)
		}
		if data.Efficiency > 0 {
			r.printKV(w, "Efficiency", fmt.Sprintf("%.1f%% of theoretical max", data.Efficiency))
		}
	}

	// Section: Latency & Packet Loss
	r.printSection(w, "Latency & Packet Loss")
	if data.GatewayPing != "" {
		r.printKV(w, "Gateway", data.GatewayPing)
	}
	if data.InternetPing != "" {
		r.printKV(w, "Internet", data.InternetPing)
	}
	if data.DNS != "" {
		r.printKV(w, "DNS", data.DNS)
	}

	// Section: DNS Configuration
	if data.DNSServersList != "" || data.DNSLeak != "" {
		r.printSection(w, "DNS Configuration")
		if data.DNSServersList != "" {
			r.printKV(w, "Servers", data.DNSServersList)
		}
		if data.DNSLeak != "" {
			r.printKV(w, "DNS Leak", data.DNSLeak)
		}
		dohStr := "No"
		if data.DoH {
			dohStr = "Yes"
		}
		dotStr := "No"
		if data.DoT {
			dotStr = "Yes"
		}
		r.printKV(w, "DoH", dohStr)
		r.printKV(w, "DoT", dotStr)
	}

	// Section: Interface Stats
	r.printSection(w, "Interface Stats")
	r.printKV(w, "TX/RX Errors", fmt.Sprintf("%d / %d", data.TxErrors, data.RxErrors))
	r.printKV(w, "Collisions", fmt.Sprintf("%d", data.Collisions))
	r.printKV(w, "MTU", fmt.Sprintf("%d", data.MTU))
	ipv6Str := "Not available"
	if data.IPv6 {
		ipv6Str = "Available"
	}
	r.printKV(w, "IPv6", ipv6Str)
	if data.Uptime != "" {
		r.printKV(w, "Uptime", data.Uptime)
	}
	captiveStr := "None"
	if data.CaptivePortal {
		captiveStr = "Detected"
		if !r.NoColor {
			captiveStr = StyleBad.Render(captiveStr)
		}
	}
	r.printKV(w, "Captive", captiveStr)
	if data.DHCPServer != "" {
		r.printKV(w, "DHCP Server", data.DHCPServer)
	}
	if data.DHCPLease != "" {
		r.printKV(w, "Lease Time", data.DHCPLease)
	}
	if data.ThroughputTx > 0 || data.ThroughputRx > 0 {
		r.printKV(w, "Throughput", fmt.Sprintf("%.1f Mbps TX / %.1f Mbps RX", data.ThroughputTx, data.ThroughputRx))
	}

	// Section: Route
	if data.TracerouteHops > 0 {
		r.printSection(w, "Route")
		r.printKV(w, "Hops", fmt.Sprintf("%d", data.TracerouteHops))
		r.printKV(w, "Final RTT", data.TracerouteFinal)
	}

	// Section: WiFi Advanced
	if data.Roaming != "" || data.DFS != "" || data.MACRandom != "" {
		r.printSection(w, "WiFi Advanced")
		if data.Roaming != "" {
			r.printKV(w, "Roaming", data.Roaming)
		}
		if data.DFS != "" {
			r.printKV(w, "DFS", data.DFS)
		}
		if data.MACRandom != "" {
			r.printKV(w, "MAC Random", data.MACRandom)
		}
	}

	// Section: Channel Analysis
	if data.ChannelCurrent > 0 {
		r.printSection(w, "Channel Analysis")
		r.printKV(w, "Current", fmt.Sprintf("%d (%s), %d neighbors", data.ChannelCurrent, data.Band, data.ChannelNeighbors))
		congestion := data.Congestion
		if !r.NoColor {
			switch congestion {
			case "Good":
				congestion = StyleGood.Render(congestion)
			case "Fair":
				congestion = StyleWarn.Render(congestion)
			default:
				congestion = StyleBad.Render(congestion)
			}
		}
		r.printKV(w, "Congestion", congestion)
		if data.BestChannel > 0 && data.BestChannel != data.ChannelCurrent {
			r.printKV(w, "Recommended", fmt.Sprintf("%d", data.BestChannel))
		}
	}

	// Section: Security
	if data.FirewallStatus != "" || data.GatewayPortsOpen != "" || data.ARPStatus != "" {
		r.printSection(w, "Security")
		if data.FirewallStatus != "" {
			r.printKV(w, "Firewall", data.FirewallStatus)
		}
		if data.GatewayPortsOpen != "" {
			r.printKV(w, "Gateway Ports", data.GatewayPortsOpen)
		}
		if data.ARPStatus != "" {
			r.printKV(w, "ARP Table", data.ARPStatus)
		}
	}

	// Section: Issues
	if len(data.Issues) > 0 {
		r.printSection(w, "Issues")
		for _, issue := range data.Issues {
			marker := "  ! "
			if !r.NoColor {
				marker = "  " + StyleWarn.Render("!") + " "
			}
			fmt.Fprintln(w, marker+issue)
		}
	}

	fmt.Fprintln(w)
	return nil
}

func (r *TableRenderer) printSection(w io.Writer, title string) {
	fmt.Fprintln(w)
	line := "── " + title + " "
	pad := 50 - len(line)
	if pad > 0 {
		line += strings.Repeat("─", pad)
	}
	if !r.NoColor {
		line = StyleDim.Render(line)
	}
	fmt.Fprintln(w, "  "+line)
}

func (r *TableRenderer) printKV(w io.Writer, key, value string) {
	label := fmt.Sprintf("  %-14s", key)
	if !r.NoColor {
		label = StyleDim.Render(label)
	}
	fmt.Fprintf(w, "%s %s\n", label, value)
}

func visibleLen(s string) int {
	// Strip ANSI escape sequences to get visible length
	result := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result++
	}
	return result
}
