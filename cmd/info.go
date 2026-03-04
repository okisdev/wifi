package cmd

import (
	"fmt"
	"os"

	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/wifi"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current WiFi connection info",
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	scanner := wifi.NewScanner()
	info, err := scanner.InterfaceInfo()
	if err != nil {
		return fmt.Errorf("failed to get interface info: %w", err)
	}

	status := "Disconnected"
	if info.Connected {
		status = "Connected"
	}

	rssiStr := ""
	if info.Connected {
		rssiStr = fmt.Sprintf("%d dBm (%s)", info.RSSI, output.SignalQuality(info.RSSI))
	}

	data := map[string]string{
		"Interface": info.Name,
		"MAC":       info.MAC,
		"SSID":      info.SSID,
		"BSSID":     info.BSSID,
		"RSSI":      rssiStr,
		"Channel":   fmt.Sprintf("%d", info.Channel),
		"Band":      string(info.Band),
		"Tx Rate":   info.TxRate,
		"Security":  info.Security,
		"IP":        info.IP,
		"Status":    status,
	}

	if info.Noise != 0 {
		data["Noise"] = fmt.Sprintf("%d dBm", info.Noise)
	}

	if wifi.MissingLocationPermission(info) {
		fmt.Fprintf(os.Stderr, "\n⚠ %s\n\n", wifi.LocationServicesWarning)
	}

	renderer := getRenderer()
	return renderer.RenderInterfaceInfo(os.Stdout, data)
}
