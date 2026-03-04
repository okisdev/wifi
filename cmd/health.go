package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/okisdev/wifi/internal/diag"
	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/speed"
	"github.com/okisdev/wifi/internal/wifi"
	"github.com/spf13/cobra"
)

var (
	noSpeed          bool
	noNetworkQuality bool
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run a comprehensive WiFi health check",
	RunE:  runHealth,
}

func init() {
	healthCmd.Flags().BoolVar(&noSpeed, "no-speed", false, "Skip speed test")
	healthCmd.Flags().BoolVar(&noNetworkQuality, "no-nq", false, "Skip networkQuality test (macOS)")
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	scanner := wifi.NewScanner()
	tester := speed.NewTester()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	opts := diag.RunOptions{
		NoSpeed:          noSpeed,
		NoNetworkQuality: noNetworkQuality,
	}

	report, err := diag.RunAll(ctx, scanner, tester, opts, func(step diag.DiagStep) {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "\r[%d/%d] %s...     ", step.Index+1, step.Total, step.Name)
		}
	})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !jsonOutput {
		fmt.Fprintln(os.Stderr) // clear progress line
	}

	data := buildHealthReportData(report)
	renderer := getRenderer()
	return renderer.RenderHealthReport(os.Stdout, data)
}

func buildHealthReportData(r *diag.HealthReport) *output.HealthReportData {
	d := &output.HealthReportData{
		Grade:         string(r.Grade),
		Score:         r.Score,
		SSID:          r.Interface.SSID,
		Channel:       r.Interface.Channel,
		Band:          string(r.Interface.Band),
		Width:         r.Interface.Width,
		Security:      r.Interface.Security,
		SecurityScore: r.SecurityScore,
		PHYMode:       r.Interface.PHYMode,
		RSSI:          r.Interface.RSSI,
		SignalQuality: output.SignalQuality(r.Interface.RSSI),
		SNR:           r.SNR,
		Efficiency:    r.Efficiency,
		MTU:           r.MTU,
		IPv6:          r.IPv6,
		CaptivePortal: r.CaptivePortal,
		Issues:        r.Issues,
	}

	if r.Speed != nil {
		d.Download = r.Speed.Download
		d.Upload = r.Speed.Upload
		d.SpeedServer = r.Speed.Server
	}

	if r.GatewayPing != nil && r.GatewayPing.Err == nil {
		d.GatewayPing = fmt.Sprintf("%.1f ms avg, %.1f ms jitter, %.1f%% loss",
			r.GatewayPing.AvgMs, r.GatewayPing.JitterMs, r.GatewayPing.LossPct)
	}

	if r.InternetPing != nil && r.InternetPing.Err == nil {
		d.InternetPing = fmt.Sprintf("%.1f ms avg, %.1f ms jitter, %.1f%% loss",
			r.InternetPing.AvgMs, r.InternetPing.JitterMs, r.InternetPing.LossPct)
	}

	if len(r.DNS) > 0 && r.DNS[0].Err == nil {
		d.DNS = fmt.Sprintf("%.0f ms (%s)", r.DNS[0].DurationMs, r.DNS[0].Resolver)
	}

	if r.IfStats != nil && r.IfStats.Err == nil {
		d.TxErrors = r.IfStats.TxErrors
		d.RxErrors = r.IfStats.RxErrors
		d.Collisions = r.IfStats.Collisions
	}

	if r.Uptime > 0 {
		d.Uptime = r.Uptime.String()
	}

	if r.ChannelHealth != nil {
		d.ChannelCurrent = r.ChannelHealth.Channel
		d.ChannelNeighbors = r.ChannelHealth.NeighborCount
		d.Congestion = r.ChannelHealth.CongestionLevel
		d.BestChannel = r.ChannelHealth.BestChannel
	}

	return d
}
