package output

import "github.com/charmbracelet/lipgloss"

var (
	ColorExcellent = lipgloss.Color("#00ff00")
	ColorGood      = lipgloss.Color("#7fff00")
	ColorFair      = lipgloss.Color("#ffff00")
	ColorWeak      = lipgloss.Color("#ff7f00")
	ColorPoor      = lipgloss.Color("#ff0000")

	StyleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00bfff"))
	StyleGood  = lipgloss.NewStyle().Foreground(ColorGood)
	StyleWarn  = lipgloss.NewStyle().Foreground(ColorFair)
	StyleBad   = lipgloss.NewStyle().Foreground(ColorPoor)
	StyleDim   = lipgloss.NewStyle().Faint(true)
)

// SignalColor returns a color based on RSSI value.
func SignalColor(rssi int) lipgloss.Color {
	switch {
	case rssi >= -30:
		return ColorExcellent
	case rssi >= -50:
		return ColorGood
	case rssi >= -60:
		return ColorFair
	case rssi >= -70:
		return ColorWeak
	default:
		return ColorPoor
	}
}

// SignalQuality returns a text description of signal quality.
func SignalQuality(rssi int) string {
	switch {
	case rssi >= -30:
		return "Excellent"
	case rssi >= -50:
		return "Good"
	case rssi >= -60:
		return "Fair"
	case rssi >= -70:
		return "Weak"
	default:
		return "Poor"
	}
}

// SignalBar returns a visual bar for signal strength.
func SignalBar(rssi int, noColor bool) string {
	// Map RSSI to 0-5 bars
	bars := 0
	switch {
	case rssi >= -30:
		bars = 5
	case rssi >= -50:
		bars = 4
	case rssi >= -60:
		bars = 3
	case rssi >= -70:
		bars = 2
	case rssi >= -80:
		bars = 1
	}

	full := "█"
	empty := "░"
	bar := ""
	for i := 0; i < 5; i++ {
		if i < bars {
			bar += full
		} else {
			bar += empty
		}
	}

	if noColor {
		return bar
	}
	return lipgloss.NewStyle().Foreground(SignalColor(rssi)).Render(bar)
}
