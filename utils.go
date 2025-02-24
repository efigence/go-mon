package mon

import (
	"bytes"
	"encoding/gob"
	"math"
	"net"
	"os"
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

// Wraps unsigned 64 bit counter to 64 signed one, on zero
func WrapUint64Counter(i uint64) (o int64) {
	if i <= math.MaxInt64 {
		return int64(i)
	} else {
		return int64(i) + math.MaxInt64 + 1
	}
}

type GobTag struct {
	T map[string]string
}

func mapToGobTag(v ...map[string]string) GobTag {
	d := map[string]string{}
	for _, m := range v {
		for k, v := range m {
			d[k] = promLabelEscapeLabelValue(v)
		}
	}
	return GobTag{T: d}
}

func gobTag(g GobTag) (data []byte) {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(&g)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}
func ungobTag(data []byte) (g GobTag) {
	b := bytes.NewReader(data)
	err := gob.NewDecoder(b).Decode(&g)
	if err != nil {
		panic(err)
	}
	return g
}

var emptyGob = gobTag(GobTag{map[string]string{}})

func promLabelEscapeLabelValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return value
}
