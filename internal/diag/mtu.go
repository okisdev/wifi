package diag

import "net"

func GetMTU(ifaceName string) int {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return 0
	}
	return iface.MTU
}
