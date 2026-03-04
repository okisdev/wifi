package diag

import (
	"os/exec"
	"regexp"
	"strings"
)

// macRe matches MAC addresses in common formats (colon-separated).
var macRe = regexp.MustCompile(`at\s+([0-9a-fA-F:]{11,17})\s`)

// CheckARPSpoof checks for ARP spoofing by detecting duplicate MAC addresses
// mapped to multiple IPs.
func CheckARPSpoof() *ARPSpoofResult {
	result := &ARPSpoofResult{}

	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		result.Err = err
		result.ErrMsg = err.Error()
		return result
	}

	// macToIPs maps a MAC address to all IPs that resolve to it.
	macToIPs := make(map[string][]string)

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract IP: look for (IP) pattern.
		ip := extractIP(line)
		if ip == "" {
			continue
		}

		// Extract MAC: look for "at XX:XX:XX:XX:XX:XX" pattern.
		matches := macRe.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		mac := strings.ToLower(matches[1])

		// Skip incomplete entries.
		if mac == "(incomplete)" || mac == "ff:ff:ff:ff:ff:ff" {
			continue
		}

		macToIPs[mac] = append(macToIPs[mac], ip)
	}

	// Find duplicates: MAC addresses mapped to more than one IP.
	dupes := make(map[string][]string)
	for mac, ips := range macToIPs {
		if len(ips) > 1 {
			dupes[mac] = ips
		}
	}

	result.DuplicateMACs = dupes
	result.IsSuspicious = len(dupes) > 0

	return result
}

// extractIP extracts an IP address from an arp output line.
// It looks for the (IP) pattern common on Darwin and Linux.
func extractIP(line string) string {
	start := strings.Index(line, "(")
	end := strings.Index(line, ")")
	if start == -1 || end == -1 || end <= start+1 {
		return ""
	}
	return line[start+1 : end]
}
