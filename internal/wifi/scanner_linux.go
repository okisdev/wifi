//go:build linux

package wifi

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

type linuxScanner struct{}

func NewScanner() Scanner {
	return &linuxScanner{}
}

func (s *linuxScanner) Scan() ([]Network, error) {
	// Try iw first (more detailed, but needs root)
	if nets, err := s.scanIW(); err == nil {
		return nets, nil
	}
	// Fallback to nmcli (no root needed)
	return s.scanNMCLI()
}

func (s *linuxScanner) scanIW() ([]Network, error) {
	iface, err := s.findInterface()
	if err != nil {
		return nil, err
	}

	out, err := exec.Command("iw", "dev", iface, "scan").Output()
	if err != nil {
		return nil, fmt.Errorf("iw scan failed (try with sudo): %w", err)
	}

	return parseIWScan(string(out)), nil
}

func parseIWScan(output string) []Network {
	var nets []Network
	var cur *Network

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "BSS ") {
			if cur != nil {
				nets = append(nets, *cur)
			}
			cur = &Network{}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				cur.BSSID = strings.TrimSuffix(parts[1], "(on")
			}
			continue
		}
		if cur == nil {
			continue
		}

		switch {
		case strings.HasPrefix(line, "SSID:"):
			cur.SSID = strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		case strings.HasPrefix(line, "signal:"):
			val := strings.TrimSpace(strings.TrimPrefix(line, "signal:"))
			val = strings.TrimSuffix(val, " dBm")
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				cur.RSSI = int(f)
			}
		case strings.HasPrefix(line, "freq:"):
			val := strings.TrimSpace(strings.TrimPrefix(line, "freq:"))
			if freq, err := strconv.Atoi(val); err == nil {
				cur.Channel = ChannelFromFrequency(freq)
				cur.Band = BandFromFrequency(freq)
			}
		case strings.HasPrefix(line, "channel width:"):
			val := strings.TrimSpace(strings.TrimPrefix(line, "channel width:"))
			val = strings.Split(val, " ")[0]
			if w, err := strconv.Atoi(val); err == nil {
				cur.Width = w
			}
		case strings.Contains(line, "WPA"):
			if strings.Contains(line, "Version: 2") {
				cur.Security = SecurityWPA2
			} else {
				cur.Security = SecurityWPA
			}
		case strings.Contains(line, "WEP"):
			cur.Security = SecurityWEP
		case strings.Contains(line, "RSN"):
			cur.Security = SecurityWPA2
		}
	}
	if cur != nil {
		nets = append(nets, *cur)
	}

	// Default security for networks without explicit security info
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

func (s *linuxScanner) scanNMCLI() ([]Network, error) {
	out, err := exec.Command("nmcli", "-t", "-f", "SSID,BSSID,SIGNAL,CHAN,FREQ,SECURITY", "dev", "wifi", "list", "--rescan", "yes").Output()
	if err != nil {
		return nil, fmt.Errorf("nmcli failed: %w", err)
	}

	var nets []Network
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 6 {
			continue
		}
		// nmcli -t uses \: for colons in BSSID, so we need to rejoin
		// Actually with -t, BSSID fields contain escaped colons
		// Let's use a simpler approach
		n := Network{SSID: fields[0]}

		// BSSID is fields[1] through fields[6] for a MAC address
		if len(fields) >= 11 {
			n.BSSID = strings.Join(fields[1:7], ":")
			n.BSSID = strings.ReplaceAll(n.BSSID, "\\", "")
			if sig, err := strconv.Atoi(fields[7]); err == nil {
				// nmcli returns signal as 0-100 percentage, convert to approx dBm
				n.RSSI = percentToDBm(sig)
			}
			if ch, err := strconv.Atoi(fields[8]); err == nil {
				n.Channel = ch
				n.Band = BandFromChannel(ch)
			}
			if freq, err := strconv.Atoi(fields[9]); err == nil {
				n.Band = BandFromFrequency(freq)
			}
			sec := fields[10]
			n.Security = parseNMCLISecurity(sec)
		}
		n.Width = 20
		nets = append(nets, n)
	}
	return nets, nil
}

func parseNMCLISecurity(s string) Security {
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
	case s == "" || s == "--":
		return SecurityOpen
	default:
		return SecurityUnknown
	}
}

func percentToDBm(pct int) int {
	// Rough conversion: 100% ≈ -30 dBm, 0% ≈ -90 dBm
	return -90 + (pct * 60 / 100)
}

func (s *linuxScanner) InterfaceInfo() (*InterfaceInfo, error) {
	iface, err := s.findInterface()
	if err != nil {
		return nil, err
	}

	info := &InterfaceInfo{Name: iface}

	// Get connection info via iw
	out, err := exec.Command("iw", "dev", iface, "link").Output()
	if err == nil {
		s.parseIWLink(string(out), info)
	}

	// Get MAC and IP
	if netIf, err := net.InterfaceByName(iface); err == nil {
		info.MAC = netIf.HardwareAddr.String()
		if addrs, err := netIf.Addrs(); err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					info.IP = ipnet.IP.String()
					break
				}
			}
		}
	}

	// Compute SNR (Linux often doesn't report noise, default to -95)
	noise := info.Noise
	if noise == 0 {
		noise = -95
	}
	info.SNR = info.RSSI - noise

	return info, nil
}

func (s *linuxScanner) parseIWLink(output string, info *InterfaceInfo) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "Connected to"):
			info.Connected = true
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				info.BSSID = parts[2]
			}
		case strings.HasPrefix(line, "SSID:"):
			info.SSID = strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		case strings.HasPrefix(line, "signal:"):
			val := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "signal:")), " dBm")
			if rssi, err := strconv.Atoi(val); err == nil {
				info.RSSI = rssi
			}
		case strings.HasPrefix(line, "freq:"):
			if freq, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "freq:"))); err == nil {
				info.Channel = ChannelFromFrequency(freq)
				info.Band = BandFromFrequency(freq)
			}
		case strings.HasPrefix(line, "tx bitrate:"):
			info.TxRate = strings.TrimSpace(strings.TrimPrefix(line, "tx bitrate:"))
			// Infer PHY mode from tx bitrate info
			switch {
			case strings.Contains(info.TxRate, "HE"):
				info.PHYMode = "802.11ax"
			case strings.Contains(info.TxRate, "VHT"):
				info.PHYMode = "802.11ac"
			case strings.Contains(info.TxRate, "HT"):
				info.PHYMode = "802.11n"
			default:
				info.PHYMode = "802.11a/g"
			}
		}
	}
}

func (s *linuxScanner) CurrentRSSI() (int, error) {
	info, err := s.InterfaceInfo()
	if err != nil {
		return 0, err
	}
	if !info.Connected {
		return 0, fmt.Errorf("not connected to any network")
	}
	return info.RSSI, nil
}

func (s *linuxScanner) findInterface() (string, error) {
	out, err := exec.Command("iw", "dev").Output()
	if err != nil {
		// Try ip link as fallback
		out2, err2 := exec.Command("ip", "link", "show").Output()
		if err2 != nil {
			return "", fmt.Errorf("cannot find WiFi interface")
		}
		// Look for wl* interface
		scanner := bufio.NewScanner(strings.NewReader(string(out2)))
		for scanner.Scan() {
			line := scanner.Text()
			for _, field := range strings.Fields(line) {
				field = strings.TrimSuffix(field, ":")
				if strings.HasPrefix(field, "wl") {
					return field, nil
				}
			}
		}
		return "", fmt.Errorf("no WiFi interface found")
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Interface") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	return "", fmt.Errorf("no WiFi interface found")
}
