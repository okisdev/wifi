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
	ChannelCurrent   int      `json:"channel_current"`
	ChannelNeighbors int      `json:"channel_neighbors"`
	Congestion       string   `json:"congestion"`
	BestChannel      int      `json:"best_channel"`
	PublicIP         string   `json:"public_ip,omitempty"`
	ISP              string   `json:"isp,omitempty"`
	ASN              string   `json:"asn,omitempty"`
	GeoLocation      string   `json:"geo_location,omitempty"`
	IsVPN            bool     `json:"is_vpn,omitempty"`
	DNSServersList   string   `json:"dns_servers,omitempty"`
	DNSLeak          string   `json:"dns_leak,omitempty"`
	DoH              bool     `json:"doh,omitempty"`
	DoT              bool     `json:"dot,omitempty"`
	TracerouteHops   int      `json:"traceroute_hops,omitempty"`
	TracerouteFinal  string   `json:"traceroute_final,omitempty"`
	DHCPServer       string   `json:"dhcp_server,omitempty"`
	DHCPLease        string   `json:"dhcp_lease,omitempty"`
	Roaming          string   `json:"roaming,omitempty"`
	DFS              string   `json:"dfs,omitempty"`
	MACRandom        string   `json:"mac_random,omitempty"`
	ThroughputTx     float64  `json:"throughput_tx_mbps,omitempty"`
	ThroughputRx     float64  `json:"throughput_rx_mbps,omitempty"`
	GatewayPortsOpen string   `json:"gateway_ports_open,omitempty"`
	FirewallStatus   string   `json:"firewall_status,omitempty"`
	ARPStatus        string   `json:"arp_status,omitempty"`
	Issues           []string `json:"issues"`
}
