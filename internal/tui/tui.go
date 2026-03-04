package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/okisdev/wifi/internal/speed"
	"github.com/okisdev/wifi/internal/wifi"
)

const (
	tabScan   = 0
	tabInfo   = 1
	tabSignal = 2
	tabSpeed  = 3
	tabHealth = 4
)

type refreshMsg struct{}

type model struct {
	activeTab int
	width     int
	height    int

	scan   scanModel
	info   infoModel
	signal signalModel
	speed  speedModel
	health healthModel
}

// Run starts the interactive TUI.
func Run() error {
	scanner := wifi.NewScanner()
	tester := speed.NewTester()

	m := model{
		activeTab: tabScan,
		scan:      newScanModel(scanner),
		info:      newInfoModel(scanner),
		signal:    newSignalModel(scanner),
		speed:     newSpeedModel(tester),
		health:    newHealthModel(scanner, tester),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.scan.Init(),
		m.info.Init(),
		m.signal.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := msg.Height - 4
		m.scan.width = msg.Width - 4
		m.scan.height = contentHeight
		m.info.width = msg.Width - 4
		m.info.height = contentHeight
		m.signal.width = msg.Width - 4
		m.signal.height = contentHeight
		m.speed.width = msg.Width - 4
		m.speed.height = contentHeight
		m.health.width = msg.Width - 4
		m.health.height = contentHeight
		return m, nil

	case tea.KeyMsg:
		prevTab := m.activeTab
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Tab):
			m.activeTab = (m.activeTab + 1) % 5
		case key.Matches(msg, keys.ShiftTab):
			m.activeTab = (m.activeTab + 4) % 5
		case key.Matches(msg, keys.Tab1):
			m.activeTab = tabScan
		case key.Matches(msg, keys.Tab2):
			m.activeTab = tabInfo
		case key.Matches(msg, keys.Tab3):
			m.activeTab = tabSignal
		case key.Matches(msg, keys.Tab4):
			m.activeTab = tabSpeed
		case key.Matches(msg, keys.Tab5):
			m.activeTab = tabHealth
		case key.Matches(msg, keys.Refresh):
			// Refresh active view only
			return m, func() tea.Msg { return refreshMsg{} }
		default:
			// View-specific key → active view only
			return m.routeToActive(msg)
		}

		// Tab switched
		if m.activeTab != prevTab {
			cmd := m.handleTabSwitch()
			return m, cmd
		}
		return m, nil

	case refreshMsg:
		// Refresh active view only
		return m.routeToActive(msg)
	}

	// All other messages (spinner ticks, async results, signal ticks)
	// → route to ALL sub-models so background tasks complete
	return m.routeToAll(msg)
}

// routeToActive sends a message only to the currently active view.
func (m model) routeToActive(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.activeTab {
	case tabScan:
		m.scan, cmd = m.scan.Update(msg)
	case tabInfo:
		m.info, cmd = m.info.Update(msg)
	case tabSignal:
		m.signal, cmd = m.signal.Update(msg)
	case tabSpeed:
		m.speed, cmd = m.speed.Update(msg)
	case tabHealth:
		m.health, cmd = m.health.Update(msg)
	}
	return m, cmd
}

// routeToAll sends a message to every sub-model.
func (m model) routeToAll(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.scan, cmd = m.scan.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.info, cmd = m.info.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.signal, cmd = m.signal.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.speed, cmd = m.speed.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.health, cmd = m.health.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleTabSwitch triggers auto-refresh when switching to certain tabs.
func (m *model) handleTabSwitch() tea.Cmd {
	if m.activeTab == tabInfo {
		m.info.loading = true
		m.info.err = nil
		return tea.Batch(m.info.spinner.Tick, m.info.fetchInfo())
	}
	return nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	contentWidth := m.width - 2

	// Tab bar
	tabBar := renderTabs(m.activeTab, contentWidth)

	// Separator
	sep := lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("─", contentWidth))

	// Content
	var content string
	switch m.activeTab {
	case tabScan:
		content = m.scan.View()
	case tabInfo:
		content = m.info.View()
	case tabSignal:
		content = m.signal.View()
	case tabSpeed:
		content = m.speed.View()
	case tabHealth:
		content = m.health.View()
	}

	// Pad content to fill height
	contentHeight := m.height - 4
	contentLines := strings.Count(content, "\n")
	if contentLines < contentHeight {
		content += strings.Repeat("\n", contentHeight-contentLines)
	}

	// Status bar
	statusBar := renderStatusBar(m.activeTab, contentWidth)

	view := fmt.Sprintf("%s\n%s\n%s%s\n%s",
		tabBar, sep, content, sep, statusBar)

	return view
}
