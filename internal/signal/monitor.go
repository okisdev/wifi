package signal

import (
	"context"
	"time"

	"github.com/okisdev/wifi/internal/wifi"
)

// Monitor polls RSSI at a given interval and records to history.
type Monitor struct {
	scanner  wifi.Scanner
	history  *History
	interval time.Duration
}

// NewMonitor creates a new signal monitor.
func NewMonitor(scanner wifi.Scanner, interval time.Duration, historySize int) *Monitor {
	return &Monitor{
		scanner:  scanner,
		history:  NewHistory(historySize),
		interval: interval,
	}
}

// Sample takes a single RSSI reading and records it.
func (m *Monitor) Sample() (int, error) {
	rssi, err := m.scanner.CurrentRSSI()
	if err != nil {
		return 0, err
	}
	m.history.Add(rssi)
	return rssi, nil
}

// Watch continuously samples RSSI, calling the callback on each sample.
// It blocks until ctx is cancelled.
func (m *Monitor) Watch(ctx context.Context, cb func(rssi int, min, max int, avg, jitter float64, samples int)) error {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			rssi, err := m.Sample()
			if err != nil {
				continue
			}
			min, max, avg, jitter := m.history.Stats()
			cb(rssi, min, max, avg, jitter, m.history.Len())
		}
	}
}

// History returns the underlying history buffer.
func (m *Monitor) History() *History {
	return m.history
}
