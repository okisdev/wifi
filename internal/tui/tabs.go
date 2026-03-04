package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var tabNames = []string{"Scan", "Info", "Signal", "Speed", "Health"}

func renderTabs(activeTab int, width int) string {
	var tabs []string
	for i, name := range tabNames {
		label := name
		if i == activeTab {
			tabs = append(tabs, styleActiveTab.Render(label))
		} else {
			tabs = append(tabs, styleInactiveTab.Render(label))
		}
	}

	title := styleTitle.Render("wifi")
	row := title + "   " + strings.Join(tabs, " ")

	// Pad to width
	gap := width - lipgloss.Width(row)
	if gap > 0 {
		row += strings.Repeat(" ", gap)
	}

	return row
}

func renderStatusBar(activeTab int, width int) string {
	base := "q:quit  tab:switch  1-5:jump  r:refresh"
	extra := ""
	switch activeTab {
	case 0: // Scan
		extra = "  s:sort  b:band  ↑↓:scroll"
	case 3: // Speed
		extra = "  enter:start  esc:cancel"
	case 4: // Health
		extra = "  enter:start  esc:cancel  ↑↓:scroll"
	}
	text := base + extra

	gap := width - lipgloss.Width(text)
	if gap > 0 {
		text += strings.Repeat(" ", gap)
	}

	return styleStatusBar.Render(text)
}
