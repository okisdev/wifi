package signal

import (
	"sort"

	"github.com/okisdev/wifi/internal/wifi"
)

// ChannelStat holds stats for a single channel.
type ChannelStat struct {
	Channel  int
	Band     wifi.Band
	Networks int
	TotalRSS int // sum of RSSI values
	AvgRSSI  int
	Rating   string
}

// Analyze analyzes channel congestion from a list of networks.
func Analyze(networks []wifi.Network) []ChannelStat {
	chanMap := make(map[int]*ChannelStat)

	for _, n := range networks {
		ch := n.Channel
		if ch == 0 {
			continue
		}
		stat, ok := chanMap[ch]
		if !ok {
			stat = &ChannelStat{
				Channel: ch,
				Band:    n.Band,
			}
			chanMap[ch] = stat
		}
		stat.Networks++
		stat.TotalRSS += n.RSSI
	}

	var stats []ChannelStat
	for _, s := range chanMap {
		if s.Networks > 0 {
			s.AvgRSSI = s.TotalRSS / s.Networks
		}
		s.Rating = channelRating(s.Networks, s.AvgRSSI)
		stats = append(stats, *s)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Channel < stats[j].Channel
	})

	return stats
}

// BestChannel returns the recommended channel for a given band.
func BestChannel(networks []wifi.Network, band wifi.Band) int {
	stats := Analyze(networks)

	// Filter by band
	var candidates []ChannelStat
	for _, s := range stats {
		if s.Band == band {
			candidates = append(candidates, s)
		}
	}

	if len(candidates) == 0 {
		// Return default channel for band
		if band == wifi.Band5GHz {
			return 36
		}
		return 1
	}

	// Available non-overlapping channels
	var available map[int]bool
	if band == wifi.Band2_4GHz {
		available = map[int]bool{1: true, 6: true, 11: true}
	} else {
		available = map[int]bool{36: true, 40: true, 44: true, 48: true, 149: true, 153: true, 157: true, 161: true}
	}

	// Find least congested available channel
	best := -1
	bestScore := 999

	for ch := range available {
		score := 0
		for _, s := range candidates {
			if s.Channel == ch {
				score = s.Networks
				break
			}
		}
		if score < bestScore {
			bestScore = score
			best = ch
		}
	}

	if best == -1 {
		if band == wifi.Band5GHz {
			return 36
		}
		return 1
	}
	return best
}

func channelRating(networks int, avgRSSI int) string {
	switch {
	case networks <= 1:
		return "Good"
	case networks <= 3 || avgRSSI < -70:
		return "Fair"
	default:
		return "Congested"
	}
}
