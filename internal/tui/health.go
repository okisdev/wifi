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

type healthProgressMsg struct {
	step diag.DiagStep
}

type healthDoneMsg struct {
	report *diag.HealthReport
	err    error
}

type healthModel struct {
	report   *diag.HealthReport
	loading  bool
	step     string
	progress int
	total    int
	err      error
	spinner  spinner.Model
	cancel   context.CancelFunc
	scanner  wifi.Scanner
	tester   speed.Tester
	scroll   int
	width    int
	height   int
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
	case healthDoneMsg:
		m.loading = false
		m.report = msg.report
		m.err = msg.err
		m.cancel = nil
		m.scroll = 0
		return m, nil

	case healthProgressMsg:
		m.progress = msg.step.Index + 1
		m.total = msg.step.Total
		m.step = msg.step.Name
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !m.loading {
				m.loading = true
				m.err = nil
				m.report = nil
				m.step = "Starting..."
				m.progress = 0
				m.total = 14
				ctx, cancel := context.WithCancel(context.Background())
				m.cancel = cancel
				return m, tea.Batch(m.spinner.Tick, m.runDiag(ctx))
			}
		case "esc":
			if m.loading && m.cancel != nil {
				m.cancel()
				m.loading = false
				m.step = ""
				m.cancel = nil
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
			m.loading = true
			m.err = nil
			m.report = nil
			m.step = "Starting..."
			m.progress = 0
			ctx, cancel := context.WithCancel(context.Background())
			m.cancel = cancel
			return m, tea.Batch(m.spinner.Tick, m.runDiag(ctx))
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m healthModel) runDiag(ctx context.Context) tea.Cmd {
	scanner := m.scanner
	tester := m.tester
	return func() tea.Msg {
		type resultCh struct {
			report *diag.HealthReport
			err    error
		}
		ch := make(chan resultCh, 1)
		go func() {
			report, err := diag.RunAll(ctx, scanner, tester, diag.RunOptions{}, func(step diag.DiagStep) {
				// We can't send tea.Msg from here directly in a safe way
				// The progress will be shown via the step name in the final report
			})
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
	if m.loading {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("\n  %s %s\n", m.spinner.View(), m.step))
		if m.total > 0 {
			pct := float64(m.progress) / float64(m.total) * 100
			barWidth := 30
			filled := int(float64(barWidth) * float64(m.progress) / float64(m.total))
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
			b.WriteString(fmt.Sprintf("\n  [%s] %.0f%% (%d/%d)\n", bar, pct, m.progress, m.total))
		}
		b.WriteString("\n  Press Esc to cancel\n")
		return b.String()
	}

	if m.report == nil {
		if m.err != nil {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("\n  Error: %v\n", m.err))
			b.WriteString("\n  Press Enter or r to retry\n")
			return b.String()
		}
		return "\n  Press Enter or r to start health check\n"
	}

	return m.renderReport()
}

func (m healthModel) renderReport() string {
	r := m.report
	var b strings.Builder

	// Grade header
	gradeColor := gradeStyle(r.Grade)
	gradeStr := gradeColor.Render(fmt.Sprintf("Grade: %s (%d)", r.Grade, r.Score))
	title := styleTitle.Render("Wi-Fi Health Report") + " — " + gradeStr
	b.WriteString("\n  " + title + "\n")
	b.WriteString("  " + strings.Repeat("─", 50) + "\n")

	// Network Info
	b.WriteString(m.sectionHeader("Network Info"))
	b.WriteString(m.kv("SSID", r.Interface.SSID))
	b.WriteString(m.kv("Channel", fmt.Sprintf("%d (%s, %dMHz)", r.Interface.Channel, r.Interface.Band, r.Interface.Width)))
	b.WriteString(m.kv("Security", fmt.Sprintf("%-20s %d/100", r.Interface.Security, r.SecurityScore)))
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

	// Latency & Packet Loss
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

	// Interface Stats
	b.WriteString(m.sectionHeader("Interface Stats"))
	if r.IfStats != nil && r.IfStats.Err == nil {
		b.WriteString(m.kv("TX/RX Errors", fmt.Sprintf("%d / %d", r.IfStats.TxErrors, r.IfStats.RxErrors)))
		b.WriteString(m.kv("Collisions", fmt.Sprintf("%d", r.IfStats.Collisions)))
	}
	b.WriteString(m.kv("MTU", fmt.Sprintf("%d", r.MTU)))
	ipv6Str := "Not available"
	if r.IPv6 {
		ipv6Str = lipgloss.NewStyle().Foreground(colorGood).Render("Available")
	}
	b.WriteString(m.kv("IPv6", ipv6Str))
	if r.Uptime > 0 {
		b.WriteString(m.kv("Uptime", r.Uptime.String()))
	}
	captive := "None"
	if r.CaptivePortal {
		captive = lipgloss.NewStyle().Foreground(colorPoor).Render("Detected")
	}
	b.WriteString(m.kv("Captive", captive))

	// Channel Analysis
	if r.ChannelHealth != nil {
		b.WriteString(m.sectionHeader("Channel Analysis"))
		b.WriteString(m.kv("Current", fmt.Sprintf("%d (%s), %d neighbors",
			r.ChannelHealth.Channel, r.Interface.Band, r.ChannelHealth.NeighborCount)))
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

	// Issues
	if len(r.Issues) > 0 {
		b.WriteString(m.sectionHeader("Issues"))
		for _, issue := range r.Issues {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(colorFair).Render("!") + " " + issue + "\n")
		}
	}

	b.WriteString("\n  Press Enter or r to run again\n")

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
