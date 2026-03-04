//go:build !darwin || !cgo

package wifi

// LocationAuthStatus returns LocationStatusAuthorized on non-macOS platforms
// or when CGO is not available (e.g. cross-compilation).
func LocationAuthStatus() int {
	return LocationStatusAuthorized
}

// RequestLocationPermission is a no-op when CGO is not available.
func RequestLocationPermission() (bool, error) {
	return true, nil
}
