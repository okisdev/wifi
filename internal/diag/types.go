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

// RunOptions configures the diagnostic run.
type RunOptions struct {
	NoSpeed          bool
	NoNetworkQuality bool
}
