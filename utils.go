package mon

import (
	"os"
	"net"
	"strings"
)

func getFQDN() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	hostAddrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname
	}
	for _, addr := range hostAddrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				return hostname
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, ".")
		} else if ipv6 := addr.To16(); ipv6 != nil {
			ip, err := ipv6.MarshalText()
			if err != nil {
				return hostname
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, ".")
		}
	}
	return hostname
}