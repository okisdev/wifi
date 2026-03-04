//go:build windows

package wifi

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

type windowsScanner struct{}

func NewScanner() Scanner {
	return &windowsScanner{}
}

func (s *windowsScanner) Scan() ([]Network, error) {
	out, err := exec.Command("netsh", "wlan", "show", "networks", "mode=bssid").Output()
	if err != nil {
		return nil, fmt.Errorf("netsh failed: %w", err)
	}
	return parseNetshScan(string(out)), nil
}

func parseNetshScan(output string) []Network {
	var nets []Network
	var cur *Network

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "SSID") && !strings.HasPrefix(trimmed, "SSID ") && strings.Contains(trimmed, ":") {
			// New network: "SSID 1 : NetworkName"
			if cur != nil {
				nets = append(nets, *cur)
			}
			cur = &Network{}
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				cur.SSID = strings.TrimSpace(parts[1])
			}
			continue
		}
		if cur == nil {
			continue
		}

		if idx := strings.Index(trimmed, ":"); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			val := strings.TrimSpace(trimmed[idx+1:])
			key = strings.ToLower(key)

			switch {
			case strings.Contains(key, "bssid"):
				cur.BSSID = val
			case strings.Contains(key, "signal"):
				val = strings.TrimSuffix(val, "%")
				if pct, err := strconv.Atoi(val); err == nil {
					cur.RSSI = percentToDBmWin(pct)
				}
			case strings.Contains(key, "channel"):
				if ch, err := strconv.Atoi(val); err == nil {
					cur.Channel = ch
					cur.Band = BandFromChannel(ch)
				}
			case strings.Contains(key, "authentication") || strings.Contains(key, "auth"):
				cur.Security = parseWindowsSecurity(val)
			case strings.Contains(key, "radio type"):
				// e.g., "802.11ac" or "802.11n"
			}
		}
	}
	if cur != nil {
		nets = append(nets, *cur)
	}

	for i := range nets {
		if nets[i].Security == "" {
			nets[i].Security = SecurityOpen
		}
		if nets[i].Width == 0 {
			nets[i].Width = 20
		}
	}
	return nets
}

func parseWindowsSecurity(s string) Security {
	s = strings.ToUpper(s)
	switch {
	case strings.Contains(s, "WPA3"):
		return SecurityWPA3
	case strings.Contains(s, "WPA2"):
		return SecurityWPA2
	case strings.Contains(s, "WPA"):
		return SecurityWPA
	case strings.Contains(s, "WEP"):
		return SecurityWEP
	case strings.Contains(s, "OPEN"):
		return SecurityOpen
	default:
		return SecurityUnknown
	}
}

func percentToDBmWin(pct int) int {
	return -90 + (pct * 60 / 100)
}

func (s *windowsScanner) InterfaceInfo() (*InterfaceInfo, error) {
	out, err := exec.Command("netsh", "wlan", "show", "interfaces").Output()
	if err != nil {
		return nil, fmt.Errorf("netsh failed: %w", err)
	}

	info := &InterfaceInfo{}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(strings.ToLower(line[:idx]))
			val := strings.TrimSpace(line[idx+1:])

			switch {
			case key == "name":
				info.Name = val
			case strings.Contains(key, "physical"):
				info.MAC = val
			case key == "ssid" || strings.Contains(key, "ssid"):
				if info.SSID == "" {
					info.SSID = val
				}
			case strings.Contains(key, "bssid"):
				info.BSSID = val
			case strings.Contains(key, "signal"):
				val = strings.TrimSuffix(val, "%")
				if pct, err := strconv.Atoi(val); err == nil {
					info.RSSI = percentToDBmWin(pct)
				}
			case strings.Contains(key, "channel"):
				if ch, err := strconv.Atoi(val); err == nil {
					info.Channel = ch
					info.Band = BandFromChannel(ch)
				}
			case strings.Contains(key, "receive rate") || strings.Contains(key, "transmit rate"):
				if info.TxRate == "" {
					info.TxRate = val + " Mbps"
				}
			case strings.Contains(key, "radio type"):
				info.PHYMode = val
			case strings.Contains(key, "authentication"):
				info.Security = val
			case strings.Contains(key, "state"):
				info.Connected = strings.Contains(strings.ToLower(val), "connected")
			}
		}
	}

	// Get IP
	if info.Name != "" {
		if netIf, err := net.InterfaceByName(info.Name); err == nil {
			if addrs, err := netIf.Addrs(); err == nil {
				for _, addr := range addrs {
					if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
						info.IP = ipnet.IP.String()
						break
					}
				}
			}
		}
	}

	// Windows doesn't expose noise floor
	noise := -95
	info.Noise = noise
	info.SNR = info.RSSI - noise

	return info, nil
}

func (s *windowsScanner) CurrentRSSI() (int, error) {
	info, err := s.InterfaceInfo()
	if err != nil {
		return 0, err
	}
	if !info.Connected {
		return 0, fmt.Errorf("not connected to any network")
	}
	return info.RSSI, nil
}
