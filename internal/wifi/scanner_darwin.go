//go:build darwin

package wifi

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreWLAN -framework Foundation

#import <CoreWLAN/CoreWLAN.h>
#import <Foundation/Foundation.h>

typedef struct {
	const char *ssid;
	const char *bssid;
	long rssi;
	long noise;
	long channel;
	long band;
	long width;
	const char *security;
} CNetwork;

typedef struct {
	const char *name;
	const char *mac;
	const char *ssid;
	const char *bssid;
	long rssi;
	long noise;
	long channel;
	long band;
	const char *tx_rate;
	const char *security;
	const char *phy_mode;
	long width;
	int connected;
} CInterfaceInfo;

static const char* copyCString(NSString *s) {
	if (!s) return strdup("");
	return strdup([s UTF8String]);
}

static const char* securityString(CWNetwork *n) {
	if (n.ibss) {
		return strdup("IBSS");
	}
	// Basic heuristic: most modern networks are WPA2
	return strdup("WPA2");
}

int scanNetworks(CNetwork **out, int *count) {
	@autoreleasepool {
		CWWiFiClient *client = [CWWiFiClient sharedWiFiClient];
		CWInterface *iface = [client interface];
		if (!iface) {
			*count = 0;
			return -1;
		}

		NSError *err = nil;
		NSSet<CWNetwork *> *networks = [iface scanForNetworksWithName:nil error:&err];
		if (err || !networks) {
			*count = 0;
			return -1;
		}

		int n = (int)[networks count];
		*count = n;
		*out = (CNetwork *)malloc(sizeof(CNetwork) * n);

		int i = 0;
		for (CWNetwork *net in networks) {
			(*out)[i].ssid = copyCString(net.ssid);
			(*out)[i].bssid = copyCString(net.bssid);
			(*out)[i].rssi = net.rssiValue;
			(*out)[i].noise = net.noiseMeasurement;
			(*out)[i].channel = (long)net.wlanChannel.channelNumber;

			// Band detection
			switch (net.wlanChannel.channelBand) {
				case kCWChannelBand2GHz:
					(*out)[i].band = 2;
					break;
				case kCWChannelBand5GHz:
					(*out)[i].band = 5;
					break;
				default:
					(*out)[i].band = 0;
					break;
			}

			// Channel width
			switch (net.wlanChannel.channelWidth) {
				case kCWChannelWidth20MHz:
					(*out)[i].width = 20;
					break;
				case kCWChannelWidth40MHz:
					(*out)[i].width = 40;
					break;
				case kCWChannelWidth80MHz:
					(*out)[i].width = 80;
					break;
				case kCWChannelWidth160MHz:
					(*out)[i].width = 160;
					break;
				default:
					(*out)[i].width = 20;
					break;
			}

			(*out)[i].security = securityString(net);
			i++;
		}
		return 0;
	}
}

int getInterfaceInfo(CInterfaceInfo *info) {
	@autoreleasepool {
		CWWiFiClient *client = [CWWiFiClient sharedWiFiClient];
		CWInterface *iface = [client interface];
		if (!iface) return -1;

		info->name = copyCString(iface.interfaceName);
		info->mac = copyCString(iface.hardwareAddress);
		info->ssid = copyCString(iface.ssid);
		info->bssid = copyCString(iface.bssid);
		info->rssi = iface.rssiValue;
		info->noise = iface.noiseMeasurement;
		// Use transmitRate to detect connectivity — iface.ssid requires
		// Location Services and returns nil without permission, even when connected.
		info->connected = (iface.transmitRate > 0) ? 1 : 0;

		if (iface.wlanChannel) {
			info->channel = (long)iface.wlanChannel.channelNumber;
			switch (iface.wlanChannel.channelBand) {
				case kCWChannelBand2GHz: info->band = 2; break;
				case kCWChannelBand5GHz: info->band = 5; break;
				default: info->band = 0; break;
			}
			// Channel width
			switch (iface.wlanChannel.channelWidth) {
				case kCWChannelWidth20MHz: info->width = 20; break;
				case kCWChannelWidth40MHz: info->width = 40; break;
				case kCWChannelWidth80MHz: info->width = 80; break;
				case kCWChannelWidth160MHz: info->width = 160; break;
				default: info->width = 20; break;
			}
		}

		// PHY mode
		switch (iface.activePHYMode) {
			case kCWPHYMode11a: info->phy_mode = strdup("802.11a"); break;
			case kCWPHYMode11b: info->phy_mode = strdup("802.11b"); break;
			case kCWPHYMode11g: info->phy_mode = strdup("802.11g"); break;
			case kCWPHYMode11n: info->phy_mode = strdup("802.11n"); break;
			case kCWPHYMode11ac: info->phy_mode = strdup("802.11ac"); break;
			case kCWPHYMode11ax: info->phy_mode = strdup("802.11ax"); break;
			default: info->phy_mode = strdup("unknown"); break;
		}

		info->tx_rate = copyCString([NSString stringWithFormat:@"%.0f Mbps", iface.transmitRate]);
		info->security = strdup("WPA2");
		return 0;
	}
}

int getCurrentRSSI(long *rssi) {
	@autoreleasepool {
		CWWiFiClient *client = [CWWiFiClient sharedWiFiClient];
		CWInterface *iface = [client interface];
		if (!iface || !iface.ssid) return -1;
		*rssi = iface.rssiValue;
		return 0;
	}
}

void freeNetworks(CNetwork *nets, int count) {
	for (int i = 0; i < count; i++) {
		free((void*)nets[i].ssid);
		free((void*)nets[i].bssid);
		free((void*)nets[i].security);
	}
	free(nets);
}

void freeInterfaceInfo(CInterfaceInfo *info) {
	free((void*)info->name);
	free((void*)info->mac);
	free((void*)info->ssid);
	free((void*)info->bssid);
	free((void*)info->tx_rate);
	free((void*)info->security);
	free((void*)info->phy_mode);
}
*/
import "C"

import (
	"fmt"
	"net"
	"unsafe"
)

type darwinScanner struct{}

func NewScanner() Scanner {
	return &darwinScanner{}
}

func (s *darwinScanner) Scan() ([]Network, error) {
	var cNets *C.CNetwork
	var count C.int

	if C.scanNetworks(&cNets, &count) != 0 {
		return nil, fmt.Errorf("failed to scan networks (ensure Location Services is enabled)")
	}
	defer C.freeNetworks(cNets, count)

	nets := make([]Network, int(count))
	cSlice := unsafe.Slice(cNets, int(count))

	for i, cn := range cSlice {
		nets[i] = Network{
			SSID:     C.GoString(cn.ssid),
			BSSID:    C.GoString(cn.bssid),
			RSSI:     int(cn.rssi),
			Channel:  int(cn.channel),
			Security: Security(C.GoString(cn.security)),
			Width:    int(cn.width),
		}
		switch cn.band {
		case 2:
			nets[i].Band = Band2_4GHz
		case 5:
			nets[i].Band = Band5GHz
		default:
			nets[i].Band = BandFromChannel(nets[i].Channel)
		}
	}
	return nets, nil
}

func (s *darwinScanner) InterfaceInfo() (*InterfaceInfo, error) {
	var ci C.CInterfaceInfo
	if C.getInterfaceInfo(&ci) != 0 {
		return nil, fmt.Errorf("failed to get interface info")
	}
	defer C.freeInterfaceInfo(&ci)

	info := &InterfaceInfo{
		Name:      C.GoString(ci.name),
		MAC:       C.GoString(ci.mac),
		SSID:      C.GoString(ci.ssid),
		BSSID:     C.GoString(ci.bssid),
		RSSI:      int(ci.rssi),
		Noise:     int(ci.noise),
		Channel:   int(ci.channel),
		TxRate:    C.GoString(ci.tx_rate),
		Security:  C.GoString(ci.security),
		PHYMode:   C.GoString(ci.phy_mode),
		Width:     int(ci.width),
		Connected: ci.connected != 0,
	}
	info.SNR = info.RSSI - info.Noise

	switch ci.band {
	case 2:
		info.Band = Band2_4GHz
	case 5:
		info.Band = Band5GHz
	default:
		info.Band = BandFromChannel(info.Channel)
	}

	// Get IP address
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if iface.Name == info.Name {
				if addrs, err := iface.Addrs(); err == nil {
					for _, addr := range addrs {
						if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
							info.IP = ipnet.IP.String()
							break
						}
					}
				}
				break
			}
		}
	}

	return info, nil
}

func (s *darwinScanner) CurrentRSSI() (int, error) {
	var rssi C.long
	if C.getCurrentRSSI(&rssi) != 0 {
		return 0, fmt.Errorf("not connected to any network")
	}
	return int(rssi), nil
}
