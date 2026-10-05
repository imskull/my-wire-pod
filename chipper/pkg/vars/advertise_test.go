package vars

import (
	"net"
	"testing"
)

func TestValidateAdvertiseIP(t *testing.T) {
	addrs := []net.Addr{
		&net.IPNet{IP: net.ParseIP("192.168.1.10"), Mask: net.CIDRMask(24, 32)},
		&net.IPNet{IP: net.ParseIP("198.18.192.8"), Mask: net.CIDRMask(21, 32)},
	}
	for _, tc := range []struct {
		value string
		ok    bool
	}{
		{"192.168.1.10", true}, {" 192.168.1.10 ", true},
		{"192.168.1.99", false}, {"127.0.0.1", false}, {"0.0.0.0", false},
		{"224.0.0.251", false}, {"::1", false}, {"broken", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			ip, err := validateAdvertiseIP(tc.value, addrs)
			if (err == nil) != tc.ok {
				t.Fatalf("ip=%v err=%v; want success=%v", ip, err, tc.ok)
			}
			if tc.ok && !ip.Equal(net.ParseIP("192.168.1.10")) {
				t.Fatalf("selected wrong interface: %s", ip)
			}
		})
	}
}
