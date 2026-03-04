package cmd

import (
	"fmt"
	"os"

	"github.com/okisdev/wifi/internal/output"
	"github.com/okisdev/wifi/internal/tui"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	noColor    bool
)

var rootCmd = &cobra.Command{
	Use:           "wifi",
	Short:         "WiFi diagnostic CLI tool",
	Long:          "A cross-platform WiFi diagnostic tool for scanning networks, testing speed, and analyzing signal quality.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

func getRenderer() output.Renderer {
	if jsonOutput {
		return output.NewJSONRenderer()
	}
	return output.NewTableRenderer(noColor)
}
