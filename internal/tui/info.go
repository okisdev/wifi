package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/wifi"
)

type infoModel struct {
	info    *wifi.InterfaceInfo
	loading bool
	err     error
	spinner spinner.Model
	scanner wifi.Scanner
	width   int
	height  int
}

type infoDoneMsg struct {
	info *wifi.InterfaceInfo
	err  error
}

func newInfoModel(scanner wifi.Scanner) infoModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styleSpinner
	return infoModel{
		scanner: scanner,
		spinner: s,
	}
}

func (m infoModel) fetchInfo() tea.Cmd {
	return func() tea.Msg {
		info, err := m.scanner.InterfaceInfo()
		return infoDoneMsg{info: info, err: err}
	}
}

func (m infoModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchInfo())
}

func (m infoModel) Update(msg tea.Msg) (infoModel, tea.Cmd) {
	switch msg := msg.(type) {
	case infoDoneMsg:
		m.info = msg.info
		m.err = msg.err
		m.loading = false
		return m, nil
	case refreshMsg:
		m.loading = true
		m.err = nil
		return m, tea.Batch(m.spinner.Tick, m.fetchInfo())
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m infoModel) View() string {
	if m.loading || m.info == nil {
		if m.err != nil {
			return fmt.Sprintf("\n  Error: %v\n", m.err)
		}
		return fmt.Sprintf("\n  %s Loading interface info...\n", m.spinner.View())
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n", m.err)
	}

	info := m.info
	var b strings.Builder

	title := styleTitle.Render("WiFi Interface")
	b.WriteString("\n  " + title + "\n")
	b.WriteString("  " + strings.Repeat("─", 45) + "\n")

	status := "Disconnected"
	statusStyle := lipgloss.NewStyle().Foreground(colorPoor)
	if info.Connected {
		status = "Connected"
		statusStyle = lipgloss.NewStyle().Foreground(colorGood)
	}

	rows := []struct {
		label string
		value string
	}{
		{"Interface", info.Name},
		{"MAC", info.MAC},
		{"Status", statusStyle.Render(status)},
		{"SSID", info.SSID},
		{"BSSID", info.BSSID},
	}

	if info.Connected {
		rssiStr := fmt.Sprintf("%d dBm (%s)", info.RSSI, output.SignalQuality(info.RSSI))
		rssiStr = signalStyle(info.RSSI).Render(rssiStr)
		rows = append(rows, struct {
			label string
			value string
		}{"RSSI", rssiStr})

		if info.Noise != 0 {
			rows = append(rows, struct {
				label string
				value string
			}{"Noise", fmt.Sprintf("%d dBm", info.Noise)})
		}

		if info.SNR > 0 {
			rows = append(rows, struct {
				label string
				value string
			}{"SNR", fmt.Sprintf("%d dB", info.SNR)})
		}

		rows = append(rows,
			struct {
				label string
				value string
			}{"Channel", fmt.Sprintf("%d", info.Channel)},
			struct {
				label string
				value string
			}{"Band", string(info.Band)},
			struct {
				label string
				value string
			}{"Tx Rate", info.TxRate},
		)

		if info.PHYMode != "" {
			rows = append(rows, struct {
				label string
				value string
			}{"PHY Mode", info.PHYMode})
		}

		rows = append(rows,
			struct {
				label string
				value string
			}{"Security", info.Security},
			struct {
				label string
				value string
			}{"IP", info.IP},
		)
	}

	for _, r := range rows {
		if r.value == "" {
			continue
		}
		b.WriteString("  " + styleLabel.Render(r.label) + " " + r.value + "\n")
	}

	return b.String()
}
