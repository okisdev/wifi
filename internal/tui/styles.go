package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/okisdev/wifi/internal/diag"
	"github.com/okisdev/wifi/internal/output"
)

var (
	colorAccent    = lipgloss.Color("#00bfff")
	colorDim       = lipgloss.Color("#666666")
	colorBorder    = lipgloss.Color("#444444")
	colorActiveTab = lipgloss.Color("#00bfff")
	colorTabText   = lipgloss.Color("#999999")

	styleApp = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder)

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	styleActiveTab = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(colorActiveTab).
			Padding(0, 2)

	styleInactiveTab = lipgloss.NewStyle().
				Foreground(colorTabText).
				Padding(0, 2)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorDim)

	styleLabel = lipgloss.NewStyle().
			Foreground(colorDim).
			Width(14)

	styleValue = lipgloss.NewStyle()

	styleHeaderRow = lipgloss.NewStyle().
			Bold(true).
			Faint(true)

	styleSpinner = lipgloss.NewStyle().
			Foreground(colorAccent)

	// Reuse colors from output package
	colorExcellent = output.ColorExcellent
	colorGood      = output.ColorGood
	colorFair      = output.ColorFair
	colorWeak      = output.ColorWeak
	colorPoor      = output.ColorPoor
)

func signalStyle(rssi int) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(output.SignalColor(rssi))
}

func speedColor(mbps float64) lipgloss.Style {
	switch {
	case mbps >= 50:
		return lipgloss.NewStyle().Foreground(colorGood)
	case mbps >= 10:
		return lipgloss.NewStyle().Foreground(colorFair)
	default:
		return lipgloss.NewStyle().Foreground(colorPoor)
	}
}

func gradeStyle(grade diag.Grade) lipgloss.Style {
	switch grade {
	case diag.GradeA:
		return lipgloss.NewStyle().Bold(true).Foreground(colorExcellent)
	case diag.GradeB:
		return lipgloss.NewStyle().Bold(true).Foreground(colorGood)
	case diag.GradeC:
		return lipgloss.NewStyle().Bold(true).Foreground(colorFair)
	case diag.GradeD:
		return lipgloss.NewStyle().Bold(true).Foreground(colorWeak)
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(colorPoor)
	}
}
