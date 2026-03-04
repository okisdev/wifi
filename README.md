# wifi

A cross-platform WiFi diagnostic CLI tool. Scan networks, test speed, monitor signal quality, and diagnose connection issues — all from your terminal.

## Features

- **Network scanning** — discover nearby WiFi networks with signal strength, channel, and security info
- **Connection info** — view current interface details (SSID, BSSID, channel, PHY mode, security)
- **Signal monitoring** — real-time signal strength with visual bar and quality rating
- **Speed testing** — measure download/upload speed, latency, and jitter
- **Health check** — comprehensive diagnostics with a scored report (gateway ping, DNS, MTU, IPv6, channel congestion, and more)
- **Doctor** — check system prerequisites and permissions
- **Interactive TUI** — tab-based interface with all features in one view
- **JSON output** — machine-readable output for scripting

## Installation

### Homebrew (macOS)

```sh
brew install okisdev/tap/wifi
```

### From source

```sh
go install github.com/okisdev/wifi@latest
```

### Binary releases

Download pre-built binaries from the [Releases](https://github.com/okisdev/wifi/releases) page.

## Quick start

Launch the interactive TUI:

```sh
wifi
```

Or use individual commands:

```sh
wifi scan          # Scan nearby networks
wifi info          # Show current connection info
wifi signal        # Monitor signal strength
wifi speed         # Run a speed test
wifi health        # Run full health diagnostics
wifi doctor        # Check system prerequisites
```

## Commands

| Command | Description |
|---------|-------------|
| `wifi` | Launch interactive TUI |
| `wifi scan` | Scan for nearby WiFi networks |
| `wifi info` | Display current WiFi connection details |
| `wifi signal` | Show real-time signal strength |
| `wifi speed` | Run download/upload speed test |
| `wifi health` | Run comprehensive health diagnostics |
| `wifi doctor` | Check system prerequisites and permissions |
| `wifi version` | Show version information |

### Global flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |
| `--no-color` | Disable colored output |

### Speed flags

| Flag | Description |
|------|-------------|
| `--download-only` | Only test download speed |
| `--upload-only` | Only test upload speed |

## TUI mode

Running `wifi` without a subcommand launches an interactive terminal UI with tabs:

- **Scan** — sortable/filterable network list
- **Info** — current connection details
- **Signal** — live signal strength monitor
- **Speed** — on-demand speed test
- **Health** — full diagnostic report with progress bar

**Keyboard shortcuts:**

| Key | Action |
|-----|--------|
| `Tab` | Switch between tabs |
| `1`–`5` | Jump to specific tab |
| `r` | Refresh current tab |
| `Enter` | Start test (Speed/Health) |
| `Esc` | Cancel running test |
| `↑`/`↓` | Scroll results |
| `s` | Sort networks (Scan tab) |
| `b` | Filter by band (Scan tab) |
| `q` | Quit |

## macOS Location Services

On macOS, scanning nearby WiFi networks requires Location Services permission. The tool will detect WiFi connections without Location Services, but `wifi scan` needs the terminal app to be granted location access in **System Settings > Privacy & Security > Location Services**.

Run `wifi doctor` to check if permissions are configured correctly.

## License

MIT
