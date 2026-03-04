package signal

import "math"

// History is a ring buffer that stores RSSI samples.
type History struct {
	buf  []int
	pos  int
	size int
	full bool
}

// NewHistory creates a ring buffer with the given capacity.
func NewHistory(capacity int) *History {
	return &History{
		buf:  make([]int, capacity),
		size: capacity,
	}
}

// Add records a new RSSI sample.
func (h *History) Add(rssi int) {
	h.buf[h.pos] = rssi
	h.pos = (h.pos + 1) % h.size
	if h.pos == 0 {
		h.full = true
	}
}

// Len returns the number of recorded samples.
func (h *History) Len() int {
	if h.full {
		return h.size
	}
	return h.pos
}

// Samples returns all recorded RSSI values in order.
func (h *History) Samples() []int {
	n := h.Len()
	result := make([]int, n)
	if h.full {
		copy(result, h.buf[h.pos:])
		copy(result[h.size-h.pos:], h.buf[:h.pos])
	} else {
		copy(result, h.buf[:h.pos])
	}
	return result
}

// Stats returns min, max, average, and jitter (stddev) of recorded samples.
func (h *History) Stats() (min, max int, avg, jitter float64) {
	n := h.Len()
	if n == 0 {
		return 0, 0, 0, 0
	}

	samples := h.Samples()
	min = samples[0]
	max = samples[0]
	sum := 0

	for _, v := range samples {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	avg = float64(sum) / float64(n)

	// Standard deviation as jitter
	if n > 1 {
		var variance float64
		for _, v := range samples {
			d := float64(v) - avg
			variance += d * d
		}
		jitter = math.Sqrt(variance / float64(n))
	}

	return
}
