//go:build !darwin
package diag

import "context"

func RunNetworkQuality(_ context.Context) *NetworkQualityResult {
	return &NetworkQualityResult{Available: false}
}
