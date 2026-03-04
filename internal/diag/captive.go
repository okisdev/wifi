package diag

import (
	"context"
	"net/http"
	"time"
)

func CheckCaptivePortal(ctx context.Context) bool {
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx2, "GET", "http://connectivitycheck.gstatic.com/generate_204", nil)
	if err != nil {
		return false
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 204 = no captive portal, anything else = captive portal
	return resp.StatusCode != 204
}
