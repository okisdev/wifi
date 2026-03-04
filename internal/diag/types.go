package diag

import (
	"time"

	"github.com/okisdev/wifi/internal/speed"
	"github.com/okisdev/wifi/internal/wifi"
)

// Grade represents a health grade from A to F.
type Grade string

const (
	GradeA Grade = "A"
	GradeB Grade = "B"
	GradeC Grade = "C"
	GradeD Grade = "D"
	GradeF Grade = "F"
)

// DiagStep describes a diagnostic step for progress reporting.
type DiagStep struct {
	Index int
	Total int
	Name  string
}

// HealthReport is the complete result of a WiFi health check.
type HealthReport struct {
	Timestamp      time.Time              `json:"timestamp"`
	Interface      *wifi.InterfaceInfo    `json:"interface"`
	SNR            int                    `json:"snr"`
	GatewayIP      string                `json:"gateway_ip"`
	GatewayPing    *PingResult           `json:"gateway_ping"`
	InternetPing   *PingResult           `json:"internet_ping"`
	DNS            []*DNSResult          `json:"dns"`
	IfStats        *InterfaceStats       `json:"interface_stats"`
	Uptime         time.Duration         `json:"uptime"`
	MTU            int                   `json:"mtu"`
	IPv6           bool                  `json:"ipv6"`
	CaptivePortal  bool                  `json:"captive_portal"`
	NetworkQuality *NetworkQualityResult `json:"network_quality"`
	Speed          *speed.Result         `json:"speed"`
	Efficiency     float64               `json:"efficiency_pct"`
	SecurityScore  int                   `json:"security_score"`
	ChannelHealth  *ChannelHealth        `json:"channel_health"`
	Identity       *NetworkIdentity     `json:"identity,omitempty"`
	DNSServers     *DNSServerInfo       `json:"dns_servers,omitempty"`
	DNSLeak        *DNSLeakResult       `json:"dns_leak,omitempty"`
	DoHDoT         *DoHDoTResult        `json:"doh_dot,omitempty"`
	Traceroute     *TracerouteResult    `json:"traceroute,omitempty"`
	DHCP           *DHCPInfo            `json:"dhcp,omitempty"`
	Roaming        *RoamingSupport      `json:"roaming,omitempty"`
	DFS            *DFSStatus           `json:"dfs,omitempty"`
	MACRandom      *MACRandomization    `json:"mac_random,omitempty"`
	Throughput     *ThroughputSample    `json:"throughput,omitempty"`
	GatewayPorts   *PortScanResult      `json:"gateway_ports,omitempty"`
	Firewall       *FirewallStatus      `json:"firewall,omitempty"`
	ARPSpoof       *ARPSpoofResult      `json:"arp_spoof,omitempty"`
	Score          int                   `json:"score"`
	Grade          Grade                 `json:"grade"`
	Issues         []string              `json:"issues"`
}

// PingResult holds the result of a ping test.
type PingResult struct {
	Target   string  `json:"target"`
	AvgMs    float64 `json:"avg_ms"`
	MinMs    float64 `json:"min_ms"`
	MaxMs    float64 `json:"max_ms"`
	JitterMs float64 `json:"jitter_ms"`
	LossPct  float64 `json:"loss_pct"`
	Err      error   `json:"-"`
	ErrMsg   string  `json:"error,omitempty"`
}

// DNSResult holds DNS resolution timing.
type DNSResult struct {
	Resolver   string  `json:"resolver"`
	DurationMs float64 `json:"duration_ms"`
	Err        error   `json:"-"`
	ErrMsg     string  `json:"error,omitempty"`
}

// InterfaceStats holds TX/RX error counters.
type InterfaceStats struct {
	TxErrors   int64 `json:"tx_errors"`
	RxErrors   int64 `json:"rx_errors"`
	Collisions int64 `json:"collisions"`
	TxPackets  int64 `json:"tx_packets"`
	RxPackets  int64 `json:"rx_packets"`
	Err        error `json:"-"`
	ErrMsg     string `json:"error,omitempty"`
}

// NetworkQualityResult holds macOS networkQuality output.
type NetworkQualityResult struct {
	DLResponsiveness int     `json:"dl_responsiveness"`
	ULResponsiveness int     `json:"ul_responsiveness"`
	DLThroughput     float64 `json:"dl_throughput_mbps"`
	ULThroughput     float64 `json:"ul_throughput_mbps"`
	Available        bool    `json:"available"`
}

// ChannelHealth holds channel congestion analysis.
type ChannelHealth struct {
	Channel         int     `json:"channel"`
	NeighborCount   int     `json:"neighbor_count"`
	CongestionLevel string  `json:"congestion_level"`
	Score           float64 `json:"score"`
	BestChannel     int     `json:"best_channel"`
}

// NetworkIdentity holds public IP and ISP information.
type NetworkIdentity struct {
	PublicIP string `json:"public_ip"`
	ISP      string `json:"isp"`
	ASN      string `json:"asn"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Org      string `json:"org"`
	IsVPN    bool   `json:"is_vpn"`
	Err      error  `json:"-"`
	ErrMsg   string `json:"error,omitempty"`
}

// DNSServerInfo holds configured DNS server information.
type DNSServerInfo struct {
	Servers []string `json:"servers"`
	Source  string   `json:"source"`
	Err     error    `json:"-"`
	ErrMsg  string   `json:"error,omitempty"`
}

// DNSLeakResult holds DNS leak test results.
type DNSLeakResult struct {
	ResolverIPs []string `json:"resolver_ips"`
	IsLeaking   bool     `json:"is_leaking"`
	Err         error    `json:"-"`
	ErrMsg      string   `json:"error,omitempty"`
}

// DoHDoTResult holds DNS-over-HTTPS/TLS support results.
type DoHDoTResult struct {
	DoHSupported bool   `json:"doh_supported"`
	DoTSupported bool   `json:"dot_supported"`
	Err          error  `json:"-"`
	ErrMsg       string `json:"error,omitempty"`
}

// TracerouteHop is a single hop in a traceroute.
type TracerouteHop struct {
	Number  int     `json:"number"`
	Address string  `json:"address"`
	RTTMs   float64 `json:"rtt_ms"`
	Timeout bool    `json:"timeout"`
}

// TracerouteResult holds traceroute results.
type TracerouteResult struct {
	Target string          `json:"target"`
	Hops   []TracerouteHop `json:"hops"`
	Err    error           `json:"-"`
	ErrMsg string          `json:"error,omitempty"`
}

// DHCPInfo holds DHCP lease information.
type DHCPInfo struct {
	ServerIP  string `json:"server_ip"`
	LeaseTime string `json:"lease_time"`
	Err       error  `json:"-"`
	ErrMsg    string `json:"error,omitempty"`
}

// RoamingSupport holds 802.11r/k/v roaming support.
type RoamingSupport struct {
	Has80211r bool   `json:"has_80211r"`
	Has80211k bool   `json:"has_80211k"`
	Has80211v bool   `json:"has_80211v"`
	Err       error  `json:"-"`
	ErrMsg    string `json:"error,omitempty"`
}

// DFSStatus holds DFS channel status.
type DFSStatus struct {
	IsDFS   bool   `json:"is_dfs"`
	Channel int    `json:"channel"`
	State   string `json:"state"`
	Err     error  `json:"-"`
	ErrMsg  string `json:"error,omitempty"`
}

// MACRandomization holds MAC address randomization status.
type MACRandomization struct {
	Enabled bool   `json:"enabled"`
	Method  string `json:"method"`
	Err     error  `json:"-"`
	ErrMsg  string `json:"error,omitempty"`
}

// ThroughputSample holds interface throughput measurement.
type ThroughputSample struct {
	TxBytesPerSec float64 `json:"tx_bytes_per_sec"`
	RxBytesPerSec float64 `json:"rx_bytes_per_sec"`
	Err           error   `json:"-"`
	ErrMsg        string  `json:"error,omitempty"`
}

// PortInfo describes an open port.
type PortInfo struct {
	Port    int    `json:"port"`
	Service string `json:"service"`
}

// PortScanResult holds gateway port scan results.
type PortScanResult struct {
	Target    string     `json:"target"`
	OpenPorts []PortInfo `json:"open_ports"`
	Err       error      `json:"-"`
	ErrMsg    string     `json:"error,omitempty"`
}

// FirewallStatus holds local firewall status.
type FirewallStatus struct {
	Enabled  bool   `json:"enabled"`
	Platform string `json:"platform"`
	Err      error  `json:"-"`
	ErrMsg   string `json:"error,omitempty"`
}

// ARPSpoofResult holds ARP spoofing detection results.
type ARPSpoofResult struct {
	DuplicateMACs map[string][]string `json:"duplicate_macs,omitempty"`
	IsSuspicious  bool                `json:"is_suspicious"`
	Err           error               `json:"-"`
	ErrMsg        string              `json:"error,omitempty"`
}

// RunOptions configures the diagnostic run.
type RunOptions struct {
	NoSpeed          bool
	NoNetworkQuality bool
	NoTraceroute     bool
	NoIdentity       bool
	NoPortScan       bool
}
