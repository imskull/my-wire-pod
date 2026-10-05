package vars

import (
	"fmt"
	"net"
	"strings"
)

// validateAdvertiseIP only accepts a unicast IPv4 address assigned to this host.
// A stale DHCP address must not silently fall back to a VPN interface.
func validateAdvertiseIP(value string, addrs []net.Addr) (net.IP, error) {
	ip := net.ParseIP(strings.TrimSpace(value)).To4()
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsLoopback() {
		return nil, fmt.Errorf("advertise_ip %q is not a usable LAN IPv4 address", value)
	}
	for _, addr := range addrs {
		local, _, err := net.ParseCIDR(addr.String())
		if err == nil && local.Equal(ip) {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("advertise_ip %s is not assigned to a local interface", ip)
}

// InterfacesForIP keeps a configured LAN announcement off VPN adapters.
func InterfacesForIP(ip net.IP) ([]net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagMulticast == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			local, _, err := net.ParseCIDR(addr.String())
			if err == nil && local.Equal(ip) {
				return []net.Interface{iface}, nil
			}
		}
	}
	return nil, fmt.Errorf("no active multicast interface owns %s", ip)
}
