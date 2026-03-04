package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/okisdev/wifi/internal/wifi"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check and fix common WiFi diagnostics issues",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		fmt.Println("All checks passed. No issues found.")
		return nil
	}

	fmt.Println("Checking Location Services...")

	// Warn if running via `go run` — each build produces a different binary
	// in the go-build cache, so macOS locationd can never match a previous
	// authorization record. The binary also has no app bundle, so it won't
	// appear in the Location Services UI.
	exe, _ := os.Executable()
	if exe != "" && isGoBuildCache(exe) {
		fmt.Fprintln(os.Stderr, "  ⚠ Running from go-build cache (go run).")
		fmt.Fprintln(os.Stderr, "    Each `go run` produces a new binary with a different hash,")
		fmt.Fprintln(os.Stderr, "    so macOS cannot persist Location Services permission.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "    Use `go build -o wifi .` and run `./wifi doctor` instead.")
		fmt.Fprintln(os.Stderr, "")
	}

	status := wifi.LocationAuthStatus()
	switch status {
	case wifi.LocationStatusAuthorized, wifi.LocationStatusAuthorizedWhenInUse:
		fmt.Println("  ✓ Location Services permission is granted.")
		return nil

	case wifi.LocationStatusDenied:
		fmt.Fprintln(os.Stderr, "  ✗ Location Services permission is denied.")
		printLocationInstructions(exe)
		openLocationSettings()
		return nil

	case wifi.LocationStatusRestricted:
		fmt.Fprintln(os.Stderr, "  ✗ Location Services is restricted (possibly by MDM or parental controls).")
		return nil

	case wifi.LocationStatusNotDetermined:
		fmt.Println("  ⚠ Location Services permission not yet granted.")
		fmt.Println("  Attempting to request permission...")

		authorized, err := wifi.RequestLocationPermission()
		if err != nil || !authorized {
			fmt.Fprintln(os.Stderr, "  ⚠ Permission dialog did not appear or was not approved.")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "    macOS attributes Location Services permission to the executable.")
			fmt.Fprintln(os.Stderr, "    CLI tools without an app bundle won't appear in System Settings.")
			printLocationInstructions(exe)
			openLocationSettings()
			return nil
		}
		fmt.Println("  ✓ Location Services permission granted! WiFi SSID/BSSID data is now available.")
		return nil

	default:
		fmt.Fprintf(os.Stderr, "  ? Unknown Location Services status: %d\n", status)
		return nil
	}
}

func isGoBuildCache(exe string) bool {
	return strings.Contains(exe, "/go-build/") || strings.Contains(exe, "\\go-build\\")
}

func printLocationInstructions(exe string) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  To grant Location Services permission:")
	fmt.Fprintln(os.Stderr, "")

	if isGoBuildCache(exe) {
		// go run — the binary path changes every build
		fmt.Fprintln(os.Stderr, "    Build a stable binary first:")
		fmt.Fprintln(os.Stderr, "      go build -o wifi .")
		fmt.Fprintln(os.Stderr, "      ./wifi doctor")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "    Then approve the system dialog when it appears.")
		fmt.Fprintln(os.Stderr, "    If no dialog appears, grant permission manually:")
	} else if isHomebrew(exe) {
		fmt.Fprintln(os.Stderr, "    The installed binary should trigger a system dialog on first use.")
		fmt.Fprintln(os.Stderr, "    If it didn't appear, grant permission manually:")
	} else {
		fmt.Fprintln(os.Stderr, "    If no system dialog appeared, grant permission manually:")
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "    1. Open System Settings → Privacy & Security → Location Services")
	fmt.Fprintln(os.Stderr, "    2. Enable Location Services (if not already on)")

	// On macOS, CLI binaries may appear in the Location Services list by their
	// executable path. If the terminal app isn't listed either, the user needs
	// to run the binary once so that locationd registers it.
	if exe != "" && !isGoBuildCache(exe) {
		name := filepath.Base(exe)
		fmt.Fprintf(os.Stderr, "    3. Look for \"%s\" or your terminal app in the list and enable it\n", name)
	} else {
		fmt.Fprintln(os.Stderr, "    3. Look for your terminal app (or the wifi binary) in the list and enable it")
	}
	fmt.Fprintln(os.Stderr, "    4. Restart your terminal and run `wifi doctor` again")
}

func isHomebrew(exe string) bool {
	return strings.Contains(exe, "/homebrew/") || strings.Contains(exe, "/Cellar/") || strings.Contains(exe, "/linuxbrew/")
}

func openLocationSettings() {
	_ = exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_LocationServices").Run()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  Opening System Settings → Location Services...")
}
