package cmd

import (
	"fmt"
	"os"

	"github.com/okisdev/wifi/internal/speed"
	"github.com/spf13/cobra"
)

var (
	downloadOnly bool
	uploadOnly   bool
)

var speedCmd = &cobra.Command{
	Use:   "speed",
	Short: "Run a speed test",
	RunE:  runSpeed,
}

func init() {
	speedCmd.Flags().BoolVar(&downloadOnly, "download-only", false, "Only test download speed")
	speedCmd.Flags().BoolVar(&uploadOnly, "upload-only", false, "Only test upload speed")
	rootCmd.AddCommand(speedCmd)
}

func runSpeed(cmd *cobra.Command, args []string) error {
	tester := speed.NewTester()

	result, err := tester.Run(downloadOnly, uploadOnly, func(phase string, _ *speed.Result) {
		fmt.Fprintf(os.Stderr, "\r\033[K%s", phase)
	})
	fmt.Fprint(os.Stderr, "\r\033[K")
	if err != nil {
		return fmt.Errorf("speed test failed: %w", err)
	}

	data := map[string]string{
		"Server":   result.Server,
		"Location": result.Location,
		"Latency":  fmt.Sprintf("%.1f ms", result.Latency),
		"Jitter":   fmt.Sprintf("%.1f ms", result.Jitter),
	}

	if !uploadOnly {
		data["Download"] = fmt.Sprintf("%.2f Mbps", result.Download)
	}
	if !downloadOnly {
		data["Upload"] = fmt.Sprintf("%.2f Mbps", result.Upload)
	}

	renderer := getRenderer()
	return renderer.RenderSpeedResult(os.Stdout, data)
}
