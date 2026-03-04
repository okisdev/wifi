package wifi

// Band represents a WiFi frequency band.
type Band string

const (
	Band2_4GHz Band = "2.4GHz"
	Band5GHz   Band = "5GHz"
	Band6GHz   Band = "6GHz"
)

// Security represents a WiFi security type.
type Security string

const (
	SecurityOpen    Security = "Open"
	SecurityWEP     Security = "WEP"
	SecurityWPA     Security = "WPA"
	SecurityWPA2    Security = "WPA2"
	SecurityWPA3    Security = "WPA3"
	SecurityUnknown Security = "Unknown"
)

// Network represents a discovered WiFi network.
type Network struct {
	SSID     string   `json:"ssid"`
	BSSID    string   `json:"bssid"`
	RSSI     int      `json:"rssi"`      // dBm
	Channel  int      `json:"channel"`
	Band     Band     `json:"band"`
	Security Security `json:"security"`
	Width    int      `json:"width"` // channel width in MHz
}

// InterfaceInfo represents the current WiFi interface status.
type InterfaceInfo struct {
	Name      string `json:"name"`
	MAC       string `json:"mac"`
	SSID      string `json:"ssid"`
	BSSID     string `json:"bssid"`
	RSSI      int    `json:"rssi"`
	Noise     int    `json:"noise"`
	Channel   int    `json:"channel"`
	Band      Band   `json:"band"`
	Width     int    `json:"width"`
	TxRate    string `json:"tx_rate"`
	PHYMode   string `json:"phy_mode"`
	SNR       int    `json:"snr"`
	Security  string `json:"security"`
	IP        string `json:"ip"`
	Connected bool   `json:"connected"`
}

// ChannelFromFrequency converts a frequency in MHz to a channel number.
func ChannelFromFrequency(freq int) int {
	switch {
	case freq >= 2412 && freq <= 2484:
		if freq == 2484 {
			return 14
		}
		return (freq - 2412) / 5 + 1
	case freq >= 5170 && freq <= 5825:
		return (freq - 5000) / 5
	case freq >= 5955 && freq <= 7115:
		return (freq - 5950) / 5
	default:
		return 0
	}
}

// BandFromChannel returns the band for a given channel number.
func BandFromChannel(ch int) Band {
	switch {
	case ch >= 1 && ch <= 14:
		return Band2_4GHz
	case ch >= 32 && ch <= 177:
		return Band5GHz
	case ch >= 1 && ch <= 233:
		return Band6GHz
	default:
		return Band2_4GHz
	}
}

// LocationServicesWarning is the message shown when Location Services permission is missing.
const LocationServicesWarning = "SSID/BSSID data unavailable. Grant Location Services permission to your terminal app:\n" +
	"  System Settings → Privacy & Security → Location Services → [Terminal App] → Enable"

// MissingLocationPermission returns true if the interface is connected but
// SSID is empty, which indicates missing Location Services permission on macOS.
func MissingLocationPermission(info *InterfaceInfo) bool {
	return info != nil && info.Connected && info.SSID == ""
}

// AllSSIDsHidden returns true if there are networks but every SSID is empty.
// This almost certainly indicates a macOS Location Services permission issue
// rather than all networks actually being hidden.
func AllSSIDsHidden(nets []Network) bool {
	if len(nets) == 0 {
		return false
	}
	for _, n := range nets {
		if n.SSID != "" {
			return false
		}
	}
	return true
}

// BandFromFrequency returns the band for a given frequency.
func BandFromFrequency(freq int) Band {
	switch {
	case freq >= 2400 && freq <= 2500:
		return Band2_4GHz
	case freq >= 5000 && freq <= 5900:
		return Band5GHz
	case freq >= 5925 && freq <= 7125:
		return Band6GHz
	default:
		return Band2_4GHz
	}
}
