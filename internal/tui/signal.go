package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/wifi"
)

type signalModel struct {
	scanner wifi.Scanner
	rssi    int
	min     int
	max     int
	avg     float64
	jitter  float64
	samples []int
	ssid    string
	err     error
	width   int
	height  int
}

// signalTickMsg is a lightweight timer signal — no I/O.
type signalTickMsg time.Time

// signalSampleMsg carries the async I/O result.
type signalSampleMsg struct {
	rssi int
	ssid string
	err  error
}

func newSignalModel(scanner wifi.Scanner) signalModel {
	return signalModel{
		scanner: scanner,
	}
}

func (m signalModel) Init() tea.Cmd {
	return m.tick()
}

// tick schedules a lightweight 1-second timer. No I/O here.
func (m signalModel) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return signalTickMsg(t)
	})
}

// sample runs the scanner asynchronously in a goroutine.
func (m signalModel) sample() tea.Cmd {
	scanner := m.scanner
	return func() tea.Msg {
		info, err := scanner.InterfaceInfo()
		if err != nil {
			return signalSampleMsg{err: err}
		}
		if !info.Connected {
			return signalSampleMsg{err: fmt.Errorf("not connected")}
		}
		return signalSampleMsg{rssi: info.RSSI, ssid: info.SSID}
	}
}

func (m signalModel) Update(msg tea.Msg) (signalModel, tea.Cmd) {
	switch msg := msg.(type) {
	case signalTickMsg:
		// Timer fired → start async I/O, schedule next tick in parallel
		return m, tea.Batch(m.sample(), m.tick())

	case signalSampleMsg:
		// Async I/O result arrived → update state
		m.err = msg.err
		if msg.err == nil {
			m.rssi = msg.rssi
			m.ssid = msg.ssid
			m.samples = append(m.samples, msg.rssi)
			if len(m.samples) > 300 {
				m.samples = m.samples[len(m.samples)-300:]
			}
			m.updateStats()
		}
		return m, nil

	case refreshMsg:
		m.samples = nil
		m.err = nil
		return m, tea.Batch(m.sample(), m.tick())
	}

	return m, nil
}

func (m *signalModel) updateStats() {
	if len(m.samples) == 0 {
		return
	}
	m.min = m.samples[0]
	m.max = m.samples[0]
	sum := 0
	for _, v := range m.samples {
		if v < m.min {
			m.min = v
		}
		if v > m.max {
			m.max = v
		}
		sum += v
	}
	m.avg = float64(sum) / float64(len(m.samples))

	if len(m.samples) > 1 {
		var variance float64
		for _, v := range m.samples {
			d := float64(v) - m.avg
			variance += d * d
		}
		m.jitter = math.Sqrt(variance / float64(len(m.samples)))
	}
}

func (m signalModel) View() string {
	var b strings.Builder

	if len(m.samples) == 0 {
		if m.err != nil {
			return fmt.Sprintf("\n  Error: %v\n", m.err)
		}
		return "\n  Waiting for signal data...\n"
	}

	b.WriteString("\n")

	// Network name
	ssid := m.ssid
	if ssid == "" {
		ssid = "(unknown)"
	}
	b.WriteString("  " + styleLabel.Render("Network") + " " + ssid + "\n\n")

	// RSSI bar
	bar := output.SignalBar(m.rssi, false)
	rssiStr := signalStyle(m.rssi).Render(fmt.Sprintf("%d dBm", m.rssi))
	quality := signalStyle(m.rssi).Render(output.SignalQuality(m.rssi))
	b.WriteString("  " + styleLabel.Render("Signal") + " " + bar + " " + rssiStr + "  " + quality + "\n")

	// Stats
	b.WriteString("  " + styleLabel.Render("Range") + " " +
		fmt.Sprintf("%d to %d dBm\n", m.min, m.max))
	b.WriteString("  " + styleLabel.Render("Average") + " " +
		fmt.Sprintf("%.1f dBm\n", m.avg))
	b.WriteString("  " + styleLabel.Render("Jitter") + " " +
		fmt.Sprintf("%.1f dB\n", m.jitter))
	b.WriteString("  " + styleLabel.Render("Samples") + " " +
		fmt.Sprintf("%d\n", len(m.samples)))

	// Sparkline
	b.WriteString("\n")
	sparkline := m.renderSparkline()
	b.WriteString("  " + styleLabel.Render("History") + " " + sparkline + "\n")

	if m.err != nil {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(colorPoor).Render(
			fmt.Sprintf("Error: %v", m.err)) + "\n")
	}

	return b.String()
}

func (m signalModel) renderSparkline() string {
	blocks := []rune("▁▂▃▄▅▆▇█")

	maxSamples := m.width - 20
	if maxSamples < 10 {
		maxSamples = 40
	}
	if maxSamples > 80 {
		maxSamples = 80
	}

	samples := m.samples
	if len(samples) > maxSamples {
		samples = samples[len(samples)-maxSamples:]
	}

	if len(samples) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, rssi := range samples {
		idx := int(float64(rssi+90) / 70.0 * 7.0)
		if idx < 0 {
			idx = 0
		}
		if idx > 7 {
			idx = 7
		}
		color := output.SignalColor(rssi)
		sb.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(blocks[idx])))
	}

	return sb.String()
}
