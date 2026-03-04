package speed

import (
	"context"
	"fmt"

	"github.com/showwin/speedtest-go/speedtest"
)

type client struct{}

// NewTester creates a new speed tester.
func NewTester() Tester {
	return &client{}
}

func (c *client) Run(downloadOnly, uploadOnly bool) (*Result, error) {
	st := speedtest.New()

	serverList, err := st.FetchServers()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}

	targets, err := serverList.FindServer([]int{})
	if err != nil || len(targets) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	server := targets[0]
	ctx := context.Background()

	if err := server.PingTestContext(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping test failed: %w", err)
	}

	result := &Result{
		Server:   server.Sponsor,
		Location: fmt.Sprintf("%s, %s", server.Name, server.Country),
		Latency:  float64(server.Latency.Milliseconds()),
		Jitter:   float64(server.Jitter.Milliseconds()),
	}

	if !uploadOnly {
		if err := server.DownloadTestContext(ctx); err != nil {
			return nil, fmt.Errorf("download test failed: %w", err)
		}
		result.Download = float64(server.DLSpeed) / 1000000 // bytes/s to Mbps
	}

	if !downloadOnly {
		if err := server.UploadTestContext(ctx); err != nil {
			return nil, fmt.Errorf("upload test failed: %w", err)
		}
		result.Upload = float64(server.ULSpeed) / 1000000 // bytes/s to Mbps
	}

	return result, nil
}
