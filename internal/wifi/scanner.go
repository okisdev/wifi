package wifi

// Scanner provides WiFi scanning and interface info.
type Scanner interface {
	Scan() ([]Network, error)
	InterfaceInfo() (*InterfaceInfo, error)
	CurrentRSSI() (int, error)
}
