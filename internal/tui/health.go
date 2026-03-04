package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/okisdev/wifi/internal/diag"
	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/speed"
	"github.com/okisdev/wifi/internal/wifi"
)

type healthProgress struct {
	step   diag.DiagStep
	report *diag.HealthReport
}

type healthProgressMsg struct {
	step   diag.DiagStep
	report *diag.HealthReport
}

type healthDoneMsg struct {
	report *diag.HealthReport
	err    error
}

type healthModel struct {
	report     *diag.HealthReport
	partial    *diag.HealthReport
	loading    bool
	step       string
	progress   int
	total      int
	err        error
	spinner    spinner.Model
	cancel     context.CancelFunc
	progressCh chan healthProgress
	scanner    wifi.Scanner
	tester     speed.Tester
	scroll     int
	width      int
	height     int
}

func newHealthModel(scanner wifi.Scanner, tester speed.Tester) healthModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = styleSpinner
	return healthModel{
		scanner: scanner,
		tester:  tester,
		spinner: s,
	}
}

func (m healthModel) Init() tea.Cmd {
	return nil // Don't auto-start
}

func (m healthModel) Update(msg tea.Msg) (healthModel, tea.Cmd) {
	switch msg := msg.(type) {
	case healthProgressMsg:
		m.progress = msg.step.Index + 1
		m.total = msg.step.Total
		m.step = msg.step.Name
		m.partial = msg.report
		if m.progressCh != nil {
			return m, waitForHealthProgress(m.progressCh)
		}
		return m, nil

	case healthDoneMsg:
		m.loading = false
		m.report = msg.report
		m.partial = nil
		m.err = msg.err
		m.cancel = nil
		m.progressCh = nil
		m.scroll = 0
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !m.loading {
				return m, m.launchDiag()
			}
		case "esc":
			if m.loading && m.cancel != nil {
				m.cancel()
				m.loading = false
				m.step = ""
				m.partial = nil
				m.cancel = nil
				m.progressCh = nil
				return m, nil
			}
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
			return m, nil
		case "down", "j":
			m.scroll++
			return m, nil
		}

	case refreshMsg:
		if !m.loading {
			return m, m.launchDiag()
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *healthModel) launchDiag() tea.Cmd {
	m.loading = true
	m.err = nil
	m.report = nil
	m.partial = nil
	m.step = "Starting..."
	m.progress = 0
	m.total = 27
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	progressCh := make(chan healthProgress, 27)
	m.progressCh = progressCh
	return tea.Batch(m.spinner.Tick, m.runDiag(ctx, progressCh), waitForHealthProgress(progressCh))
}

func waitForHealthProgress(ch chan healthProgress) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return healthProgressMsg{step: p.step, report: p.report}
	}
}

func (m healthModel) runDiag(ctx context.Context, progressCh chan healthProgress) tea.Cmd {
	scanner := m.scanner
	tester := m.tester
	return func() tea.Msg {
		type resultCh struct {
			report *diag.HealthReport
			err    error
		}
		ch := make(chan resultCh, 1)
		go func() {
			report, err := diag.RunAll(ctx, scanner, tester, diag.RunOptions{}, func(step diag.DiagStep, partial *diag.HealthReport) {
				select {
				case progressCh <- healthProgress{step: step, report: partial}:
				default:
				}
			})
			close(progressCh)
			ch <- resultCh{report: report, err: err}
		}()

		select {
		case <-ctx.Done():
			return healthDoneMsg{err: fmt.Errorf("cancelled")}
		case res := <-ch:
			return healthDoneMsg{report: res.report, err: res.err}
		}
	}
}

func (m healthModel) View() string {
	if !m.loading && m.report == nil {
		if m.err != nil {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("\n  Error: %v\n", m.err))
			b.WriteString("\n  Press Enter or r to retry\n")
			return b.String()
		}
		return "\n  Press Enter or r to start health check\n"
	}

	// Render final or partial report
	r := m.report
	if r == nil {
		r = m.partial
	}

	return m.renderReport(r)
}

func (m healthModel) renderReport(r *diag.HealthReport) string {
	var b strings.Builder

	// Grade header (only when complete)
	if !m.loading && r != nil && r.Grade != "" {
		gradeColor := gradeStyle(r.Grade)
		gradeStr := gradeColor.Render(fmt.Sprintf("Grade: %s (%d)", r.Grade, r.Score))
		title := styleTitle.Render("Wi-Fi Health Report") + " — " + gradeStr
		b.WriteString("\n  " + title + "\n")
	} else {
		title := styleTitle.Render("Wi-Fi Health Report")
		b.WriteString("\n  " + title + "\n")
	}
	b.WriteString("  " + strings.Repeat("─", 50) + "\n")

	if r == nil {
		// No partial data yet
		if m.loading {
			b.WriteString(fmt.Sprintf("\n  %s %s\n", m.spinner.View(), m.step))
			b.WriteString("\n  Press Esc to cancel\n")
		}
		return b.String()
	}

	// Network Info
	if r.Interface != nil {
		b.WriteString(m.sectionHeader("Network Info"))
		b.WriteString(m.kv("SSID", r.Interface.SSID))
		b.WriteString(m.kv("Channel", fmt.Sprintf("%d (%s, %dMHz)", r.Interface.Channel, r.Interface.Band, r.Interface.Width)))
		if r.SecurityScore > 0 {
			b.WriteString(m.kv("Security", fmt.Sprintf("%-20s %d/100", r.Interface.Security, r.SecurityScore)))
		} else {
			b.WriteString(m.kv("Security", r.Interface.Security))
		}
		if r.Interface.PHYMode != "" {
			b.WriteString(m.kv("PHY Mode", r.Interface.PHYMode))
		}

		rssiStr := signalStyle(r.Interface.RSSI).Render(
			fmt.Sprintf("%d dBm (%s)", r.Interface.RSSI, output.SignalQuality(r.Interface.RSSI)))
		bar := output.SignalBar(r.Interface.RSSI, false)
		b.WriteString(m.kv("Signal", rssiStr+"  "+bar))
		if r.SNR > 0 {
			b.WriteString(m.kv("SNR", fmt.Sprintf("%d dB", r.SNR)))
		}
	}

	// Network Identity
	if r.Identity != nil && r.Identity.Err == nil {
		b.WriteString(m.sectionHeader("Network Identity"))
		b.WriteString(m.kv("Public IP", r.Identity.PublicIP))
		if r.Identity.ISP != "" {
			b.WriteString(m.kv("ISP", r.Identity.ISP))
		}
		if r.Identity.ASN != "" {
			b.WriteString(m.kv("ASN", r.Identity.ASN))
		}
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
		if loc != "" {
			b.WriteString(m.kv("Location", loc))
		}
		if r.Identity.IsVPN {
			vpn := lipgloss.NewStyle().Foreground(colorFair).Render("Detected")
			b.WriteString(m.kv("VPN/Relay", vpn))
		}
	}

	// Speed Test
	if r.Speed != nil {
		b.WriteString(m.sectionHeader("Speed Test"))
		b.WriteString(m.kv("Server", r.Speed.Server))
		dlColor := speedColor(r.Speed.Download)
		b.WriteString(m.kv("Download", dlColor.Render(fmt.Sprintf("%.2f Mbps", r.Speed.Download))))
		ulColor := speedColor(r.Speed.Upload)
		b.WriteString(m.kv("Upload", ulColor.Render(fmt.Sprintf("%.2f Mbps", r.Speed.Upload))))
		if r.Efficiency > 0 {
			b.WriteString(m.kv("Efficiency", fmt.Sprintf("%.1f%% of theoretical max", r.Efficiency)))
		}
	}

	// Latency & Packet Loss — show section if any data available
	hasLatency := (r.GatewayPing != nil && r.GatewayPing.Err == nil) ||
		(r.InternetPing != nil && r.InternetPing.Err == nil) ||
		(len(r.DNS) > 0 && r.DNS[0].Err == nil)
	if hasLatency {
		b.WriteString(m.sectionHeader("Latency & Packet Loss"))
		if r.GatewayPing != nil && r.GatewayPing.Err == nil {
			b.WriteString(m.kv("Gateway", fmt.Sprintf("%.1f ms avg, %.1f ms jitter, %.1f%% loss",
				r.GatewayPing.AvgMs, r.GatewayPing.JitterMs, r.GatewayPing.LossPct)))
		}
		if r.InternetPing != nil && r.InternetPing.Err == nil {
			b.WriteString(m.kv("Internet", fmt.Sprintf("%.1f ms avg, %.1f ms jitter, %.1f%% loss",
				r.InternetPing.AvgMs, r.InternetPing.JitterMs, r.InternetPing.LossPct)))
		}
		if len(r.DNS) > 0 && r.DNS[0].Err == nil {
			b.WriteString(m.kv("DNS", fmt.Sprintf("%.0f ms (%s)", r.DNS[0].DurationMs, r.DNS[0].Resolver)))
		}
	}

	// DNS Configuration
	hasDNSConfig := (r.DNSServers != nil && r.DNSServers.Err == nil && len(r.DNSServers.Servers) > 0) ||
		r.DNSLeak != nil || r.DoHDoT != nil
	if hasDNSConfig {
		b.WriteString(m.sectionHeader("DNS Configuration"))
		if r.DNSServers != nil && r.DNSServers.Err == nil && len(r.DNSServers.Servers) > 0 {
			b.WriteString(m.kv("Servers", strings.Join(r.DNSServers.Servers, ", ")))
		}
		if r.DNSLeak != nil && r.DNSLeak.Err == nil {
			leakStr := lipgloss.NewStyle().Foreground(colorGood).Render("No leak")
			if r.DNSLeak.IsLeaking {
				leakStr = lipgloss.NewStyle().Foreground(colorPoor).Render("Leak detected")
			}
			b.WriteString(m.kv("DNS Leak", leakStr))
		}
		if r.DoHDoT != nil && r.DoHDoT.Err == nil {
			dohStr := "No"
			if r.DoHDoT.DoHSupported {
				dohStr = lipgloss.NewStyle().Foreground(colorGood).Render("Yes")
			}
			dotStr := "No"
			if r.DoHDoT.DoTSupported {
				dotStr = lipgloss.NewStyle().Foreground(colorGood).Render("Yes")
			}
			b.WriteString(m.kv("DoH", dohStr))
			b.WriteString(m.kv("DoT", dotStr))
		}
	}

	// Interface Stats — show section if any data available
	hasStats := (r.IfStats != nil && r.IfStats.Err == nil) || r.MTU > 0 || r.Uptime > 0
	if hasStats {
		b.WriteString(m.sectionHeader("Interface Stats"))
		if r.IfStats != nil && r.IfStats.Err == nil {
			b.WriteString(m.kv("TX/RX Errors", fmt.Sprintf("%d / %d", r.IfStats.TxErrors, r.IfStats.RxErrors)))
			b.WriteString(m.kv("Collisions", fmt.Sprintf("%d", r.IfStats.Collisions)))
		}
		if r.MTU > 0 {
			b.WriteString(m.kv("MTU", fmt.Sprintf("%d", r.MTU)))
		}
		// IPv6 only meaningful after it's been checked (step 8+)
		if r.MTU > 0 { // MTU is step 7, IPv6 is step 8; if MTU is set, IPv6 check has run
			ipv6Str := "Not available"
			if r.IPv6 {
				ipv6Str = lipgloss.NewStyle().Foreground(colorGood).Render("Available")
			}
			b.WriteString(m.kv("IPv6", ipv6Str))
		}
		if r.Uptime > 0 {
			b.WriteString(m.kv("Uptime", r.Uptime.String()))
		}
		if r.CaptivePortal {
			captive := lipgloss.NewStyle().Foreground(colorPoor).Render("Detected")
			b.WriteString(m.kv("Captive", captive))
		}
		if r.DHCP != nil && r.DHCP.Err == nil {
			if r.DHCP.ServerIP != "" {
				b.WriteString(m.kv("DHCP Server", r.DHCP.ServerIP))
			}
			if r.DHCP.LeaseTime != "" {
				b.WriteString(m.kv("Lease Time", r.DHCP.LeaseTime))
			}
		}
		if r.Throughput != nil && r.Throughput.Err == nil {
			txMbps := r.Throughput.TxBytesPerSec * 8 / 1_000_000
			rxMbps := r.Throughput.RxBytesPerSec * 8 / 1_000_000
			b.WriteString(m.kv("Throughput", fmt.Sprintf("%.1f Mbps TX / %.1f Mbps RX", txMbps, rxMbps)))
		}
	}

	// Route
	if r.Traceroute != nil && r.Traceroute.Err == nil && len(r.Traceroute.Hops) > 0 {
		b.WriteString(m.sectionHeader("Route"))
		hopCount := len(r.Traceroute.Hops)
		lastHop := r.Traceroute.Hops[hopCount-1]
		lastRTT := "timeout"
		if !lastHop.Timeout {
			lastRTT = fmt.Sprintf("%.1f ms", lastHop.RTTMs)
		}
		b.WriteString(m.kv("Target", r.Traceroute.Target))
		b.WriteString(m.kv("Hops", fmt.Sprintf("%d", hopCount)))
		b.WriteString(m.kv("Final RTT", lastRTT))
	}

	// WiFi Advanced
	hasAdvanced := r.Roaming != nil || r.DFS != nil || r.MACRandom != nil
	if hasAdvanced {
		b.WriteString(m.sectionHeader("WiFi Advanced"))
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
			roamStr := "None detected"
			if len(features) > 0 {
				roamStr = strings.Join(features, ", ")
			}
			b.WriteString(m.kv("Roaming", roamStr))
		}
		if r.DFS != nil && r.DFS.Err == nil {
			dfsStr := "No"
			if r.DFS.IsDFS {
				dfsStr = fmt.Sprintf("Yes (ch %d, %s)", r.DFS.Channel, r.DFS.State)
			}
			b.WriteString(m.kv("DFS", dfsStr))
		}
		if r.MACRandom != nil && r.MACRandom.Err == nil {
			macStr := "Disabled (hardware)"
			if r.MACRandom.Enabled {
				macStr = lipgloss.NewStyle().Foreground(colorGood).Render("Enabled (locally-administered)")
			}
			b.WriteString(m.kv("MAC Random", macStr))
		}
	}

	// Channel Analysis
	if r.ChannelHealth != nil {
		b.WriteString(m.sectionHeader("Channel Analysis"))
		band := ""
		if r.Interface != nil {
			band = string(r.Interface.Band)
		}
		b.WriteString(m.kv("Current", fmt.Sprintf("%d (%s), %d neighbors",
			r.ChannelHealth.Channel, band, r.ChannelHealth.NeighborCount)))
		congestion := r.ChannelHealth.CongestionLevel
		switch congestion {
		case "Good":
			congestion = lipgloss.NewStyle().Foreground(colorGood).Render(congestion)
		case "Fair":
			congestion = lipgloss.NewStyle().Foreground(colorFair).Render(congestion)
		default:
			congestion = lipgloss.NewStyle().Foreground(colorPoor).Render(congestion)
		}
		b.WriteString(m.kv("Congestion", congestion))
		if r.ChannelHealth.BestChannel > 0 && r.ChannelHealth.BestChannel != r.ChannelHealth.Channel {
			b.WriteString(m.kv("Recommended", fmt.Sprintf("%d", r.ChannelHealth.BestChannel)))
		}
	}

	// Security
	hasSecurity := r.GatewayPorts != nil || r.Firewall != nil || r.ARPSpoof != nil
	if hasSecurity {
		b.WriteString(m.sectionHeader("Security"))
		if r.Firewall != nil && r.Firewall.Err == nil {
			fwStr := lipgloss.NewStyle().Foreground(colorPoor).Render("Disabled")
			if r.Firewall.Enabled {
				fwStr = lipgloss.NewStyle().Foreground(colorGood).Render("Enabled")
			}
			b.WriteString(m.kv("Firewall", fmt.Sprintf("%s (%s)", fwStr, r.Firewall.Platform)))
		}
		if r.GatewayPorts != nil && r.GatewayPorts.Err == nil {
			if len(r.GatewayPorts.OpenPorts) == 0 {
				b.WriteString(m.kv("Gateway Ports", "None open"))
			} else {
				ports := []string{}
				for _, p := range r.GatewayPorts.OpenPorts {
					ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Service))
				}
				b.WriteString(m.kv("Gateway Ports", strings.Join(ports, ", ")))
			}
		}
		if r.ARPSpoof != nil && r.ARPSpoof.Err == nil {
			arpStr := lipgloss.NewStyle().Foreground(colorGood).Render("OK")
			if r.ARPSpoof.IsSuspicious {
				arpStr = lipgloss.NewStyle().Foreground(colorPoor).Render("Suspicious — duplicate MACs detected")
			}
			b.WriteString(m.kv("ARP Table", arpStr))
		}
	}

	// Issues
	if len(r.Issues) > 0 {
		b.WriteString(m.sectionHeader("Issues"))
		for _, issue := range r.Issues {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(colorFair).Render("!") + " " + issue + "\n")
		}
	}

	// Footer: loading spinner or action hint
	if m.loading {
		b.WriteString(fmt.Sprintf("\n  %s %s", m.spinner.View(), m.step))
		if m.total > 0 {
			pct := float64(m.progress) / float64(m.total) * 100
			b.WriteString(fmt.Sprintf("  (%.0f%%)", pct))
		}
		b.WriteString("\n\n  Press Esc to cancel\n")
	} else {
		b.WriteString("\n  Press Enter or r to run again\n")
	}

	// Apply scrolling
	lines := strings.Split(b.String(), "\n")
	maxScroll := len(lines) - m.height + 2
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}
	if m.scroll > 0 && m.scroll < len(lines) {
		lines = lines[m.scroll:]
	}

	return strings.Join(lines, "\n")
}

func (m healthModel) sectionHeader(title string) string {
	line := "── " + title + " "
	pad := 50 - len(line)
	if pad > 0 {
		line += strings.Repeat("─", pad)
	}
	return "\n  " + lipgloss.NewStyle().Faint(true).Render(line) + "\n"
}

func (m healthModel) kv(key, value string) string {
	label := lipgloss.NewStyle().Foreground(colorDim).Width(16).Render(key)
	return "  " + label + " " + value + "\n"
}
