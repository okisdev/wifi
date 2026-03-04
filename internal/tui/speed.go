package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/okisdev/wifi/internal/speed"
)

type speedModel struct {
	result  *speed.Result
	partial *speed.Result
	loading bool
	phase   string
	err     error
	spinner spinner.Model
	tester  speed.Tester
	cancel  context.CancelFunc
	phaseCh chan speedProgress
	width   int
	height  int
}

type speedProgress struct {
	phase   string
	partial *speed.Result
}

type speedPhaseMsg speedProgress

type speedDoneMsg struct {
	result *speed.Result
	err    error
}

func newSpeedModel(tester speed.Tester) speedModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = styleSpinner
	return speedModel{
		tester:  tester,
		spinner: s,
	}
}

func (m speedModel) Init() tea.Cmd {
	return nil // Don't auto-start
}

func (m speedModel) Update(msg tea.Msg) (speedModel, tea.Cmd) {
	switch msg := msg.(type) {
	case speedPhaseMsg:
		m.phase = msg.phase
		m.partial = msg.partial
		if m.phaseCh != nil {
			return m, waitForSpeedPhase(m.phaseCh)
		}
		return m, nil

	case speedDoneMsg:
		m.loading = false
		m.result = msg.result
		m.partial = nil
		m.err = msg.err
		m.cancel = nil
		m.phaseCh = nil
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !m.loading {
				return m, m.launchTest()
			}
		case "esc":
			if m.loading && m.cancel != nil {
				m.cancel()
				m.loading = false
				m.phase = ""
				m.partial = nil
				m.cancel = nil
				m.phaseCh = nil
				return m, nil
			}
		}

	case refreshMsg:
		if !m.loading {
			return m, m.launchTest()
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *speedModel) launchTest() tea.Cmd {
	m.loading = true
	m.err = nil
	m.result = nil
	m.partial = nil
	m.phase = "Starting..."
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	phaseCh := make(chan speedProgress, 4)
	m.phaseCh = phaseCh
	return tea.Batch(m.spinner.Tick, m.startTest(ctx, phaseCh), waitForSpeedPhase(phaseCh))
}

func waitForSpeedPhase(ch chan speedProgress) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return speedPhaseMsg(p)
	}
}

func (m speedModel) startTest(ctx context.Context, phaseCh chan speedProgress) tea.Cmd {
	tester := m.tester
	return func() tea.Msg {
		type resultCh struct {
			result *speed.Result
			err    error
		}
		ch := make(chan resultCh, 1)
		go func() {
			r, err := tester.Run(false, false, func(phase string, partial *speed.Result) {
				select {
				case phaseCh <- speedProgress{phase: phase, partial: partial}:
				default:
				}
			})
			close(phaseCh)
			ch <- resultCh{result: r, err: err}
		}()

		select {
		case <-ctx.Done():
			return speedDoneMsg{err: fmt.Errorf("cancelled")}
		case res := <-ch:
			return speedDoneMsg{result: res.result, err: res.err}
		}
	}
}

func (m speedModel) View() string {
	var b strings.Builder

	if !m.loading && m.result == nil {
		if m.err != nil {
			b.WriteString(fmt.Sprintf("\n  Error: %v\n", m.err))
			b.WriteString("\n  Press Enter or r to retry\n")
			return b.String()
		}
		b.WriteString("\n  Press Enter or r to start speed test\n")
		return b.String()
	}

	// Show result (final or partial)
	r := m.result
	if r == nil {
		r = m.partial
	}

	title := styleTitle.Render("Speed Test Results")
	b.WriteString("\n  " + title + "\n")
	b.WriteString("  " + strings.Repeat("─", 45) + "\n")

	if r != nil {
		if r.Server != "" {
			b.WriteString("  " + styleLabel.Render("Server") + " " + r.Server + "\n")
		}
		if r.Location != "" {
			b.WriteString("  " + styleLabel.Render("Location") + " " + r.Location + "\n")
		}
		if r.Latency > 0 {
			b.WriteString("  " + styleLabel.Render("Latency") + " " + fmt.Sprintf("%.1f ms", r.Latency) + "\n")
		}
		if r.Jitter > 0 {
			b.WriteString("  " + styleLabel.Render("Jitter") + " " + fmt.Sprintf("%.1f ms", r.Jitter) + "\n")
		}
		if r.Download > 0 {
			b.WriteString("  " + styleLabel.Render("Download") + " " + speedColor(r.Download).Render(fmt.Sprintf("%.2f Mbps", r.Download)) + "\n")
		}
		if r.Upload > 0 {
			b.WriteString("  " + styleLabel.Render("Upload") + " " + speedColor(r.Upload).Render(fmt.Sprintf("%.2f Mbps", r.Upload)) + "\n")
		}
	}

	if m.loading {
		b.WriteString(fmt.Sprintf("\n  %s %s\n", m.spinner.View(), m.phase))
		b.WriteString("\n  Press Esc to cancel\n")
	} else {
		if m.err != nil {
			b.WriteString("\n  " + lipgloss.NewStyle().Foreground(colorPoor).Render(
				fmt.Sprintf("Error: %v", m.err)) + "\n")
		}
		b.WriteString("\n  Press Enter or r to run again\n")
	}

	return b.String()
}
