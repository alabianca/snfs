package util

import (
	"fmt"
	"net"
	"strings"
)

// MyIP returns the local non loopback ip address
// given the network (ipv4 | ipv6)
func MyIP(network string) (net.IP, error) {

	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		var ip net.IP

		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP

		case *net.IPAddr:
			ip = v.IP
		}

		if ip.IsLoopback() {
			continue
		}

		isIPV6 := strings.Contains(ip.String(), ":")
		if isIPV6 && network == "ipv6" {
			return ip, nil
		}

		if !isIPV6 && network == "ipv4" {
			return ip, nil
		}

	}

	return nil, fmt.Errorf("No IP found")

}
