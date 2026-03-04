package diag

import (
	"strconv"
	"strings"
)

// phyMaxMbps returns the theoretical maximum throughput for a PHY mode and channel width.
func phyMaxMbps(phyMode string, widthMHz int) float64 {
	mode := strings.ToLower(phyMode)
	switch {
	case strings.Contains(mode, "ax") || strings.Contains(mode, "wifi 6"):
		switch widthMHz {
		case 160:
			return 9608
		case 80:
			return 4804
		case 40:
			return 2402
		default:
			return 1201
		}
	case strings.Contains(mode, "ac") || strings.Contains(mode, "wifi 5"):
		switch widthMHz {
		case 160:
			return 6933
		case 80:
			return 3467
		case 40:
			return 1733
		default:
			return 867
		}
	case strings.Contains(mode, "11n") || strings.Contains(mode, "wifi 4"):
		switch widthMHz {
		case 40:
			return 600
		default:
			return 300
		}
	case strings.Contains(mode, "11g"):
		return 54
	case strings.Contains(mode, "11b"):
		return 11
	case strings.Contains(mode, "11a"):
		return 54
	default:
		return 0
	}
}

func CalcEfficiency(dlMbps float64, phyMode string, widthMHz int, txRate string) float64 {
	maxMbps := phyMaxMbps(phyMode, widthMHz)

	// If PHY max is unknown, try to use txRate
	if maxMbps == 0 && txRate != "" {
		rate := strings.TrimSuffix(strings.TrimSpace(txRate), "Mbps")
		rate = strings.TrimSpace(rate)
		if r, err := strconv.ParseFloat(rate, 64); err == nil && r > 0 {
			maxMbps = r
		}
	}

	if maxMbps == 0 || dlMbps == 0 {
		return 0
	}
	eff := (dlMbps / maxMbps) * 100
	if eff > 100 {
		eff = 100
	}
	return eff
}
