package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/okisdev/wifi/internal/output"
	sig "github.com/okisdev/wifi/internal/signal"
	"github.com/okisdev/wifi/internal/wifi"
	"github.com/spf13/cobra"
)

var (
	watchMode bool
	analyze   bool
	interval  time.Duration
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Monitor WiFi signal strength",
	RunE:  runSignal,
}

func init() {
	signalCmd.Flags().BoolVar(&watchMode, "watch", false, "Continuously monitor signal (Ctrl+C to stop)")
	signalCmd.Flags().BoolVar(&analyze, "analyze", false, "Analyze channel congestion")
	signalCmd.Flags().DurationVarP(&interval, "interval", "i", time.Second, "Polling interval for --watch mode")
	rootCmd.AddCommand(signalCmd)
}

func runSignal(cmd *cobra.Command, args []string) error {
	scanner := wifi.NewScanner()

	if analyze {
		return runAnalyze(scanner)
	}

	if watchMode {
		return runWatch(scanner)
	}

	return runSingleSignal(scanner)
}

func runSingleSignal(scanner wifi.Scanner) error {
	info, err := scanner.InterfaceInfo()
	if err != nil {
		return fmt.Errorf("failed to get interface info: %w", err)
	}
	if !info.Connected {
		return fmt.Errorf("not connected to any WiFi network")
	}

	if wifi.MissingLocationPermission(info) {
		fmt.Fprintf(os.Stderr, "⚠ %s\n\n", wifi.LocationServicesWarning)
	}

	data := output.SignalData{
		SSID:    info.SSID,
		RSSI:    info.RSSI,
		Min:     info.RSSI,
		Max:     info.RSSI,
		Avg:     float64(info.RSSI),
		Jitter:  0,
		Samples: 1,
	}

	renderer := getRenderer()
	return renderer.RenderSignal(os.Stdout, data)
}

func runWatch(scanner wifi.Scanner) error {
	info, err := scanner.InterfaceInfo()
	if err != nil {
		return fmt.Errorf("failed to get interface info: %w", err)
	}
	if !info.Connected {
		return fmt.Errorf("not connected to any WiFi network")
	}

	if wifi.MissingLocationPermission(info) {
		fmt.Fprintf(os.Stderr, "⚠ %s\n\n", wifi.LocationServicesWarning)
	}

	monitor := sig.NewMonitor(scanner, interval, 300)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	renderer := getRenderer()
	ssid := info.SSID

	fmt.Fprintf(os.Stderr, "Monitoring signal for '%s' (Ctrl+C to stop)...\n\n", ssid)

	return monitor.Watch(ctx, func(rssi int, min, max int, avg, jitter float64, samples int) {
		if !jsonOutput {
			// Clear previous output with ANSI escape
			fmt.Print("\033[6A\033[J")
		}

		data := output.SignalData{
			SSID:    ssid,
			RSSI:    rssi,
			Min:     min,
			Max:     max,
			Avg:     avg,
			Jitter:  jitter,
			Samples: samples,
		}
		renderer.RenderSignal(os.Stdout, data)
	})
}

func runAnalyze(scanner wifi.Scanner) error {
	fmt.Fprintln(os.Stderr, "Scanning networks for channel analysis...")

	networks, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	stats := sig.Analyze(networks)

	channels := make([]output.ChannelInfo, len(stats))
	for i, s := range stats {
		channels[i] = output.ChannelInfo{
			Channel:  s.Channel,
			Band:     string(s.Band),
			Networks: s.Networks,
			AvgRSSI:  s.AvgRSSI,
			Rating:   s.Rating,
		}
	}

	renderer := getRenderer()
	if err := renderer.RenderChannelAnalysis(os.Stdout, channels); err != nil {
		return err
	}

	// Recommend best channels
	best24 := sig.BestChannel(networks, wifi.Band2_4GHz)
	best5 := sig.BestChannel(networks, wifi.Band5GHz)
	fmt.Fprintf(os.Stdout, "\nRecommended channels: 2.4GHz → %d, 5GHz → %d\n", best24, best5)

	return nil
}
