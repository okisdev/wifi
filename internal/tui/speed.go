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
	loading bool
	phase   string
	err     error
	spinner spinner.Model
	tester  speed.Tester
	cancel  context.CancelFunc
	width   int
	height  int
}

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
	case speedDoneMsg:
		m.loading = false
		m.result = msg.result
		m.err = msg.err
		m.cancel = nil
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !m.loading {
				m.loading = true
				m.err = nil
				m.result = nil
				m.phase = "Finding best server..."
				ctx, cancel := context.WithCancel(context.Background())
				m.cancel = cancel
				return m, tea.Batch(m.spinner.Tick, m.startTest(ctx))
			}
		case "esc":
			if m.loading && m.cancel != nil {
				m.cancel()
				m.loading = false
				m.phase = ""
				m.cancel = nil
				return m, nil
			}
		}

	case refreshMsg:
		if !m.loading {
			m.loading = true
			m.err = nil
			m.result = nil
			m.phase = "Finding best server..."
			ctx, cancel := context.WithCancel(context.Background())
			m.cancel = cancel
			return m, tea.Batch(m.spinner.Tick, m.startTest(ctx))
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m speedModel) startTest(ctx context.Context) tea.Cmd {
	tester := m.tester
	return func() tea.Msg {
		type resultCh struct {
			result *speed.Result
			err    error
		}
		ch := make(chan resultCh, 1)
		go func() {
			r, err := tester.Run(false, false)
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

	if m.loading {
		b.WriteString(fmt.Sprintf("\n  %s %s\n", m.spinner.View(), m.phase))
		b.WriteString("\n  Press Esc to cancel\n")
		return b.String()
	}

	if m.result == nil {
		if m.err != nil {
			b.WriteString(fmt.Sprintf("\n  Error: %v\n", m.err))
			b.WriteString("\n  Press Enter or r to retry\n")
			return b.String()
		}
		b.WriteString("\n  Press Enter or r to start speed test\n")
		return b.String()
	}

	r := m.result
	title := styleTitle.Render("Speed Test Results")
	b.WriteString("\n  " + title + "\n")
	b.WriteString("  " + strings.Repeat("─", 45) + "\n")

	rows := []struct {
		label string
		value string
	}{
		{"Server", r.Server},
		{"Location", r.Location},
		{"Latency", fmt.Sprintf("%.1f ms", r.Latency)},
		{"Jitter", fmt.Sprintf("%.1f ms", r.Jitter)},
		{"Download", speedColor(r.Download).Render(fmt.Sprintf("%.2f Mbps", r.Download))},
		{"Upload", speedColor(r.Upload).Render(fmt.Sprintf("%.2f Mbps", r.Upload))},
	}

	for _, row := range rows {
		b.WriteString("  " + styleLabel.Render(row.label) + " " + row.value + "\n")
	}

	b.WriteString("\n  Press Enter or r to run again\n")

	if m.err != nil {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(colorPoor).Render(
			fmt.Sprintf("Error: %v", m.err)) + "\n")
	}

	return b.String()
}
