package diag

import (
	"fmt"
	"strconv"
	"strings"
)

// CheckMACRandomization checks whether a MAC address is randomized.
// A locally-administered MAC has the 0x02 bit set in the first octet.
func CheckMACRandomization(mac string) *MACRandomization {
	result := &MACRandomization{}

	sep := ":"
	if strings.Contains(mac, "-") {
		sep = "-"
	}

	parts := strings.Split(mac, sep)
	if len(parts) < 1 || parts[0] == "" {
		result.Err = fmt.Errorf("invalid MAC address: %s", mac)
		result.ErrMsg = result.Err.Error()
		return result
	}

	firstOctet, err := strconv.ParseUint(parts[0], 16, 8)
	if err != nil {
		result.Err = fmt.Errorf("parse first octet %q: %w", parts[0], err)
		result.ErrMsg = result.Err.Error()
		return result
	}

	if firstOctet&0x02 != 0 {
		result.Enabled = true
		result.Method = "locally-administered"
	} else {
		result.Enabled = false
		result.Method = "hardware"
	}

	return result
}
