package speed

// Result holds speed test results.
type Result struct {
	Server   string  `json:"server"`
	Location string  `json:"location"`
	Latency  float64 `json:"latency_ms"`
	Jitter   float64 `json:"jitter_ms"`
	Download float64 `json:"download_mbps"`
	Upload   float64 `json:"upload_mbps"`
}

// Tester runs speed tests.
type Tester interface {
	Run(downloadOnly, uploadOnly bool, progressFn func(phase string, partial *Result)) (*Result, error)
}
