package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade wifi to the latest version",
	RunE:  runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

const githubRepo = "okisdev/wifi"

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	// Detect if managed by Homebrew
	if isHomebrew(exe) {
		fmt.Println("Installed via Homebrew. Run:")
		fmt.Println("  brew upgrade wifi")
		return nil
	}

	// Detect if running from go-build cache
	if isGoBuildCache(exe) {
		fmt.Println("Running from go-build cache (go run).")
		fmt.Println("Install a released binary first:")
		fmt.Println("  brew install okisdev/tap/wifi")
		fmt.Println("  # or download from https://github.com/okisdev/wifi/releases")
		return nil
	}

	fmt.Printf("Current version: %s\n", version)
	fmt.Println("Checking for updates...")

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version, "v")

	if current == latest {
		fmt.Printf("Already up to date (v%s).\n", latest)
		return nil
	}

	if current == "dev" {
		fmt.Printf("Latest version: v%s (you are on dev build)\n", latest)
	} else {
		fmt.Printf("New version available: v%s → v%s\n", current, latest)
	}

	// Find matching asset
	assetName := expectedAssetName(latest)
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (%s)\nDownload manually: https://github.com/%s/releases/tag/%s",
			runtime.GOOS, runtime.GOARCH, assetName, githubRepo, release.TagName)
	}

	fmt.Printf("Downloading %s...\n", assetName)
	binary, err := downloadAndExtract(downloadURL, assetName)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Replace the current binary
	if err := replaceBinary(exe, binary); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("Upgraded to v%s\n", latest)
	return nil
}

func fetchLatestRelease() (*ghRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func expectedAssetName(ver string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("wifi_%s_%s_%s.%s", ver, goos, goarch, ext)
}

func downloadAndExtract(url, assetName string) ([]byte, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(assetName, ".zip") {
		return extractFromZip(data)
	}
	return extractFromTarGz(data)
}

func extractFromTarGz(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if filepath.Base(hdr.Name) == "wifi" && hdr.Typeflag == tar.TypeReg {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("wifi binary not found in archive")
}

func extractFromZip(data []byte) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if name == "wifi" || name == "wifi.exe" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("wifi binary not found in archive")
}

func replaceBinary(target string, newBinary []byte) error {
	// Get original file permissions
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	mode := info.Mode()

	// Write to temp file in same directory (ensures same filesystem for rename)
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, "wifi-upgrade-*")
	if err != nil {
		return fmt.Errorf("cannot create temp file (do you have write permission to %s?): %w", dir, err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(newBinary); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, mode); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, target); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot replace binary (try running with sudo): %w", err)
	}

	return nil
}
