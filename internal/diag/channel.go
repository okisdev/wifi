package diag

import (
	"github.com/okisdev/wifi/internal/signal"
	"github.com/okisdev/wifi/internal/wifi"
)

func AnalyzeChannel(currentChannel int, band wifi.Band, networks []wifi.Network) *ChannelHealth {
	stats := signal.Analyze(networks)

	ch := &ChannelHealth{
		Channel: currentChannel,
	}

	// Count neighbors on same channel
	for _, s := range stats {
		if s.Channel == currentChannel {
			ch.NeighborCount = s.Networks - 1 // exclude self
			if ch.NeighborCount < 0 {
				ch.NeighborCount = 0
			}
			break
		}
	}

	// Congestion level
	switch {
	case ch.NeighborCount <= 1:
		ch.CongestionLevel = "Good"
		ch.Score = 100
	case ch.NeighborCount <= 3:
		ch.CongestionLevel = "Fair"
		ch.Score = 60
	case ch.NeighborCount <= 5:
		ch.CongestionLevel = "Busy"
		ch.Score = 30
	default:
		ch.CongestionLevel = "Congested"
		ch.Score = 10
	}

	ch.BestChannel = signal.BestChannel(networks, band)
	return ch
}
