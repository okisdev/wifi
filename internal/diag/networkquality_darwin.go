//go:build darwin
package diag

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"
)

type nqJSON struct {
	DLFlows          int     `json:"dl_flows"`
	ULFlows          int     `json:"ul_flows"`
	DLThroughput     float64 `json:"dl_throughput"`
	ULThroughput     float64 `json:"ul_throughput"`
	DLResponsiveness int     `json:"dl_responsiveness"`
	ULResponsiveness int     `json:"ul_responsiveness"`
	Responsiveness   int     `json:"responsiveness"`
}

func RunNetworkQuality(ctx context.Context) *NetworkQualityResult {
	result := &NetworkQualityResult{Available: true}

	ctx2, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx2, "networkQuality", "-c", "-s").Output()
	if err != nil {
		result.Available = false
		return result
	}

	var nq nqJSON
	if err := json.Unmarshal(out, &nq); err != nil {
		result.Available = false
		return result
	}

	result.DLResponsiveness = nq.DLResponsiveness
	result.ULResponsiveness = nq.ULResponsiveness
	result.DLThroughput = nq.DLThroughput / 1_000_000 // bps to Mbps
	result.ULThroughput = nq.ULThroughput / 1_000_000
	return result
}
