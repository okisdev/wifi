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

// ComputeSecurityScore extends the base protocol score with penalties from
// port scan, firewall, and ARP spoof results.
func ComputeSecurityScore(security string, report *HealthReport) int {
	score := ScoreSecurity(security)

	// Penalty: insecure ports open on gateway
	if report.GatewayPorts != nil && report.GatewayPorts.Err == nil {
		for _, p := range report.GatewayPorts.OpenPorts {
			switch p.Port {
			case 23, 21, 161: // Telnet, FTP, SNMP
				score -= 10
			}
		}
	}

	// Penalty: firewall disabled
	if report.Firewall != nil && report.Firewall.Err == nil && !report.Firewall.Enabled {
		score -= 10
	}

	// Penalty: ARP spoof detected
	if report.ARPSpoof != nil && report.ARPSpoof.Err == nil && report.ARPSpoof.IsSuspicious {
		score -= 15
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}
