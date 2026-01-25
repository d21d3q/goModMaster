package netutil

import (
	"net"
	"sort"
	"strings"
)

func DiscoverIPv4() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{"localhost"}
	}

	addresses := []string{"localhost"}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagPointToPoint != 0 {
			continue
		}
		name := strings.ToLower(iface.Name)
		if strings.HasPrefix(name, "utun") || strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "lo") {
			continue
		}
		if !isPhysicalInterface(name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := extractIPv4(addr)
			if ip == "" {
				continue
			}
			addresses = append(addresses, ip)
		}
	}

	sort.Strings(addresses)
	return uniqueStrings(addresses)
}

func isPhysicalInterface(name string) bool {
	for _, prefix := range []string{"en", "eth", "wlan", "wl"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func extractIPv4(addr net.Addr) string {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil {
		return ""
	}
	ip = ip.To4()
	if ip == nil {
		return ""
	}
	return ip.String()
}

func uniqueStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
