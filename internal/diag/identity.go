package diag

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// vpnKeywords are substrings that hint the connection is through a VPN or
// privacy relay when found inside the org / ISP name returned by ipinfo.io.
var vpnKeywords = []string{
	"VPN",
	"Private",
	"Relay",
	"Tunnel",
	"NordVPN",
	"ExpressVPN",
	"Mullvad",
	"Surfshark",
	"ProtonVPN",
	"Cloudflare Warp",
}

// ipInfoResponse mirrors the subset of the ipinfo.io JSON we care about.
type ipInfoResponse struct {
	IP      string `json:"ip"`
	City    string `json:"city"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

// CheckIdentity queries ipinfo.io to determine the public IP, ISP, ASN,
// approximate location, and whether the connection appears to traverse a VPN.
func CheckIdentity(ctx context.Context) *NetworkIdentity {
	id := &NetworkIdentity{}

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx2, "GET", "https://ipinfo.io/json", nil)
	if err != nil {
		id.Err = err
		id.ErrMsg = fmt.Sprintf("build request: %v", err)
		return id
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		id.Err = err
		id.ErrMsg = fmt.Sprintf("ipinfo request: %v", err)
		return id
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		id.Err = fmt.Errorf("ipinfo returned status %d", resp.StatusCode)
		id.ErrMsg = id.Err.Error()
		return id
	}

	var info ipInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		id.Err = err
		id.ErrMsg = fmt.Sprintf("decode ipinfo response: %v", err)
		return id
	}

	id.PublicIP = info.IP
	id.City = info.City
	id.Country = info.Country
	id.Org = info.Org

	// The org field from ipinfo.io has the form "AS12345 Company Name".
	// Split it into ASN and ISP name.
	if info.Org != "" {
		parts := strings.SplitN(info.Org, " ", 2)
		if len(parts) >= 1 {
			id.ASN = parts[0]
		}
		if len(parts) >= 2 {
			id.ISP = parts[1]
		}
	}

	// Heuristic VPN detection: check whether the org string contains any
	// well-known VPN / privacy-relay keywords (case-insensitive).
	orgUpper := strings.ToUpper(info.Org)
	for _, kw := range vpnKeywords {
		if strings.Contains(orgUpper, strings.ToUpper(kw)) {
			id.IsVPN = true
			break
		}
	}

	return id
}
