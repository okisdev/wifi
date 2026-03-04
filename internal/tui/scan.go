package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/wifi"
)

type scanTickMsg time.Time

type sortMode int

const (
	sortBySignal sortMode = iota
	sortByName
	sortByChannel
)

func (s sortMode) String() string {
	switch s {
	case sortByName:
		return "Name"
	case sortByChannel:
		return "Channel"
	default:
		return "Signal"
	}
}

type bandFilter int

const (
	bandAll bandFilter = iota
	band2_4
	band5
	band6
)

func (b bandFilter) String() string {
	switch b {
	case band2_4:
		return "2.4GHz"
	case band5:
		return "5GHz"
	case band6:
		return "6GHz"
	default:
		return "All"
	}
}

type scanModel struct {
	networks   []wifi.Network
	filtered   []wifi.Network
	loading    bool
	err        error
	spinner    spinner.Model
	scanner    wifi.Scanner
	sortBy     sortMode
	bandFilter bandFilter
	cursor     int
	offset     int
	width      int
	height     int
	prevRSSI   map[string]int // BSSID → previous RSSI for trend display
}

type scanDoneMsg struct {
	networks []wifi.Network
	err      error
}

func newScanModel(scanner wifi.Scanner) scanModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styleSpinner
	return scanModel{
		scanner: scanner,
		spinner: s,
		loading: true,
	}
}

func (m scanModel) doScan() tea.Cmd {
	return func() tea.Msg {
		networks, err := m.scanner.Scan()
		return scanDoneMsg{networks: networks, err: err}
	}
}

func (m scanModel) tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return scanTickMsg(t)
	})
}

func (m scanModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.doScan(), m.tick())
}

func (m *scanModel) applyFilterSort() {
	// Filter
	var filtered []wifi.Network
	for _, n := range m.networks {
		switch m.bandFilter {
		case band2_4:
			if n.Band != wifi.Band2_4GHz {
				continue
			}
		case band5:
			if n.Band != wifi.Band5GHz {
				continue
			}
		case band6:
			if n.Band != wifi.Band6GHz {
				continue
			}
		}
		filtered = append(filtered, n)
	}

	// Sort
	switch m.sortBy {
	case sortByName:
		sort.Slice(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].SSID) < strings.ToLower(filtered[j].SSID)
		})
	case sortByChannel:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Channel < filtered[j].Channel
		})
	default:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].RSSI > filtered[j].RSSI
		})
	}

	m.filtered = filtered
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m scanModel) Update(msg tea.Msg) (scanModel, tea.Cmd) {
	switch msg := msg.(type) {
	case scanTickMsg:
		// Silent background refresh — no loading state to avoid flicker
		return m, tea.Batch(m.doScan(), m.tick())

	case scanDoneMsg:
		// Snapshot current RSSI before replacing for trend column
		if len(m.networks) > 0 {
			m.prevRSSI = make(map[string]int, len(m.networks))
			for _, n := range m.networks {
				m.prevRSSI[n.BSSID] = n.RSSI
			}
		}
		m.networks = msg.networks
		m.err = msg.err
		m.loading = false
		m.applyFilterSort()
		return m, nil

	case refreshMsg:
		m.loading = true
		m.err = nil
		return m, tea.Batch(m.spinner.Tick, m.doScan(), m.tick())

	case tea.KeyMsg:
		switch {
		case msg.String() == "s":
			m.sortBy = (m.sortBy + 1) % 3
			m.applyFilterSort()
			return m, nil
		case msg.String() == "b":
			m.bandFilter = (m.bandFilter + 1) % 4
			m.applyFilterSort()
			return m, nil
		case msg.String() == "up" || msg.String() == "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
			return m, nil
		case msg.String() == "down" || msg.String() == "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				visibleRows := m.visibleRows()
				if m.cursor >= m.offset+visibleRows {
					m.offset = m.cursor - visibleRows + 1
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m scanModel) visibleRows() int {
	// height minus header(2) + status line(1) + filter info(1) + padding(2)
	rows := m.height - 6
	if rows < 3 {
		rows = 3
	}
	return rows
}

func (m scanModel) View() string {
	if m.loading {
		if m.err != nil {
			return fmt.Sprintf("\n  Error: %v\n", m.err)
		}
		return fmt.Sprintf("\n  %s Scanning networks...\n", m.spinner.View())
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n", m.err)
	}

	var b strings.Builder

	// Filter/sort status
	filterInfo := fmt.Sprintf("  %d networks  |  Sort: %s  |  Band: %s",
		len(m.filtered), m.sortBy.String(), m.bandFilter.String())
	b.WriteString(styleStatusBar.Render(filterInfo) + "\n")

	// Location Services warning
	if wifi.AllSSIDsHidden(m.networks) {
		warning := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFCC00")).
			Render("  ⚠ " + wifi.LocationServicesWarning)
		b.WriteString(warning + "\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString("\n  No networks found.\n")
		return b.String()
	}

	// Table header
	header := fmt.Sprintf("  %-24s %-17s %-8s %-6s %-4s %-6s %-10s %-6s %s",
		"SSID", "BSSID", "SIGNAL", "TREND", "CH", "BAND", "SECURITY", "WIDTH", "QUALITY")
	b.WriteString(styleHeaderRow.Render(header) + "\n")

	// Rows
	visibleRows := m.visibleRows()
	end := m.offset + visibleRows
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := m.offset; i < end; i++ {
		n := m.filtered[i]
		ssid := n.SSID
		if ssid == "" {
			ssid = "(hidden)"
		}
		if len(ssid) > 23 {
			ssid = ssid[:20] + "..."
		}

		rssiStr := fmt.Sprintf("%d dBm", n.RSSI)
		quality := output.SignalQuality(n.RSSI)
		bar := output.SignalBar(n.RSSI, false)
		color := output.SignalColor(n.RSSI)

		rssiStr = lipgloss.NewStyle().Foreground(color).Render(rssiStr)
		quality = lipgloss.NewStyle().Foreground(color).Render(quality)

		// Trend: compare with previous scan
		trend := "  ·   "
		if prev, ok := m.prevRSSI[n.BSSID]; ok {
			delta := n.RSSI - prev
			switch {
			case delta > 0:
				trend = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(fmt.Sprintf(" +%-3d ", delta))
			case delta < 0:
				trend = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Render(fmt.Sprintf(" %-4d ", delta))
			default:
				trend = "  ·   "
			}
		}

		row := fmt.Sprintf("  %-24s %-17s %-8s %s %-4d %-6s %-10s %-6s %s %s",
			ssid, n.BSSID, rssiStr, trend, n.Channel, string(n.Band),
			string(n.Security), fmt.Sprintf("%dMHz", n.Width),
			bar, quality)

		if i == m.cursor {
			row = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Render(row)
		}

		b.WriteString(row + "\n")
	}

	// Scroll indicator
	if len(m.filtered) > visibleRows {
		b.WriteString(fmt.Sprintf("\n  %s",
			styleStatusBar.Render(fmt.Sprintf("showing %d-%d of %d", m.offset+1, end, len(m.filtered)))))
	}

	return b.String()
}
