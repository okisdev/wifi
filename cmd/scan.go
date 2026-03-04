package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/wifi"
	"github.com/spf13/cobra"
)

var (
	scanSort string
	scanBand string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for nearby WiFi networks",
	RunE:  runScan,
}

func init() {
	scanCmd.Flags().StringVar(&scanSort, "sort", "signal", "Sort by: signal, name, channel")
	scanCmd.Flags().StringVar(&scanBand, "band", "", "Filter by band: 2.4, 5, 6")
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	scanner := wifi.NewScanner()

	fmt.Fprintln(os.Stderr, "Scanning for networks...")
	networks, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Filter by band
	if scanBand != "" {
		var filtered []wifi.Network
		for _, n := range networks {
			match := false
			switch scanBand {
			case "2.4":
				match = n.Band == wifi.Band2_4GHz
			case "5":
				match = n.Band == wifi.Band5GHz
			case "6":
				match = n.Band == wifi.Band6GHz
			}
			if match {
				filtered = append(filtered, n)
			}
		}
		networks = filtered
	}

	// Sort
	switch strings.ToLower(scanSort) {
	case "name":
		sort.Slice(networks, func(i, j int) bool {
			return strings.ToLower(networks[i].SSID) < strings.ToLower(networks[j].SSID)
		})
	case "channel":
		sort.Slice(networks, func(i, j int) bool {
			return networks[i].Channel < networks[j].Channel
		})
	default: // signal
		sort.Slice(networks, func(i, j int) bool {
			return networks[i].RSSI > networks[j].RSSI
		})
	}

	// Convert to display rows
	rows := make([]output.NetworkRow, len(networks))
	for i, n := range networks {
		rows[i] = output.NetworkRow{
			SSID:     n.SSID,
			BSSID:    n.BSSID,
			Signal:   n.RSSI,
			Channel:  n.Channel,
			Band:     string(n.Band),
			Security: string(n.Security),
			Width:    n.Width,
			Bar:      output.SignalBar(n.RSSI, noColor),
		}
	}

	renderer := getRenderer()
	return renderer.RenderNetworks(os.Stdout, rows)
}
