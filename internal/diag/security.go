package diag

import "strings"

func ScoreSecurity(security string) int {
	sec := strings.ToUpper(security)
	switch {
	case strings.Contains(sec, "WPA3"):
		return 100
	case strings.Contains(sec, "WPA2"):
		return 75
	case strings.Contains(sec, "WPA"):
		return 40
	case strings.Contains(sec, "WEP"):
		return 15
	case sec == "" || strings.Contains(sec, "OPEN") || strings.Contains(sec, "NONE"):
		return 0
	default:
		return 50
	}
}
