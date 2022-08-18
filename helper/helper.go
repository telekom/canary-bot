package helper

import (
	"errors"
	"hash/fnv"
	"log"
	"net"
	"regexp"
)

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			log.Printf("Lookup Outbound IP res: %+v\n", ip)
			return ip.String(), nil
		}
	}
	return "", errors.New("Could not get outbound IP. Are you connected to the network?")
}

func LookupIP(url string) (string, error) {
	ips, err := net.LookupIP(url)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", nil
	}
	return ips[0].String(), nil
}

func LookupAddress(ip string) (string, error) {
	addr, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}
	if len(addr) == 0 {
		return "", nil
	}
	return addr[0], nil
}

func ValidateAddress(domain string) bool {
	RegExp := regexp.MustCompile(`^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z]{2,3})$`)
	return RegExp.MatchString(domain)
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
