package diag

import (
	"context"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

// commonPorts maps port numbers to their well-known service names.
var commonPorts = map[int]string{
	21:   "FTP",
	22:   "SSH",
	23:   "Telnet",
	53:   "DNS",
	80:   "HTTP",
	161:  "SNMP",
	443:  "HTTPS",
	445:  "SMB",
	8080: "HTTP-Alt",
	8443: "HTTPS-Alt",
}

// ScanGatewayPorts scans common TCP ports on the given gateway IP.
func ScanGatewayPorts(ctx context.Context, gatewayIP string) *PortScanResult {
	result := &PortScanResult{
		Target: gatewayIP,
	}

	ports := make([]int, 0, len(commonPorts))
	for p := range commonPorts {
		ports = append(ports, p)
	}

	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		sem  = make(chan struct{}, 5)
	)

	for _, port := range ports {
		if ctx.Err() != nil {
			result.Err = ctx.Err()
			result.ErrMsg = ctx.Err().Error()
			return result
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }()

			if ctx.Err() != nil {
				return
			}

			addr := gatewayIP + ":" + strconv.Itoa(p)
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				return
			}
			conn.Close()

			mu.Lock()
			result.OpenPorts = append(result.OpenPorts, PortInfo{
				Port:    p,
				Service: commonPorts[p],
			})
			mu.Unlock()
		}(port)
	}

	wg.Wait()

	sort.Slice(result.OpenPorts, func(i, j int) bool {
		return result.OpenPorts[i].Port < result.OpenPorts[j].Port
	})

	return result
}
