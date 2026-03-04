package output

import "io"

// Renderer can render data as table or JSON.
type Renderer interface {
	RenderNetworks(w io.Writer, nets []NetworkRow) error
	RenderInterfaceInfo(w io.Writer, info map[string]string) error
	RenderSpeedResult(w io.Writer, result map[string]string) error
	RenderSignal(w io.Writer, data SignalData) error
	RenderChannelAnalysis(w io.Writer, channels []ChannelInfo) error
	RenderHealthReport(w io.Writer, data *HealthReportData) error
}

// NetworkRow is a display row for network scan.
type NetworkRow struct {
	SSID     string
	BSSID    string
	Signal   int
	Channel  int
	Band     string
	Security string
	Width    int
	Bar      string
}

// SignalData holds signal display data.
type SignalData struct {
	SSID    string
	RSSI    int
	Min     int
	Max     int
	Avg     float64
	Jitter  float64
	Samples int
	Bar     string
}

// ChannelInfo holds per-channel analysis data.
type ChannelInfo struct {
	Channel  int
	Band     string
	Networks int
	AvgRSSI  int
	Rating   string
}

// HealthReportData holds display data for health reports.
type HealthReportData struct {
	Grade          string  `json:"grade"`
	Score          int     `json:"score"`
	SSID           string  `json:"ssid"`
	Channel        int     `json:"channel"`
	Band           string  `json:"band"`
	Width          int     `json:"width"`
	Security       string  `json:"security"`
	SecurityScore  int     `json:"security_score"`
	PHYMode        string  `json:"phy_mode"`
	RSSI           int     `json:"rssi"`
	SignalQuality  string  `json:"signal_quality"`
	SNR            int     `json:"snr"`
	Download       float64 `json:"download_mbps"`
	Upload         float64 `json:"upload_mbps"`
	SpeedServer    string  `json:"speed_server"`
	Efficiency     float64 `json:"efficiency_pct"`
	GatewayPing    string  `json:"gateway_ping"`
	InternetPing   string  `json:"internet_ping"`
	DNS            string  `json:"dns"`
	TxErrors       int64   `json:"tx_errors"`
	RxErrors       int64   `json:"rx_errors"`
	Collisions     int64   `json:"collisions"`
	MTU            int     `json:"mtu"`
	IPv6           bool    `json:"ipv6"`
	Uptime         string  `json:"uptime"`
	CaptivePortal  bool    `json:"captive_portal"`
	ChannelCurrent int     `json:"channel_current"`
	ChannelNeighbors int   `json:"channel_neighbors"`
	Congestion     string  `json:"congestion"`
	BestChannel    int     `json:"best_channel"`
	Issues         []string `json:"issues"`
}
