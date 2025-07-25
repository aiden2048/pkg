package maddr

import (
	"fmt"
	"net"
	"strings"
)

var (
	privateBlocks []*net.IPNet
)

func init() {
	privateBlocks = genIpBlocks([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10", "fd00::/8"})
}

func genIpBlocks(ipBlocks []string) []*net.IPNet {
	blocks := make([]*net.IPNet, 0, len(ipBlocks))
	for _, b := range ipBlocks {
		if strings.Index(b, "/") < 0 {
			b = b + "/32"
		}
		if _, block, err := net.ParseCIDR(b); err == nil {
			blocks = append(blocks, block)
		}
	}
	return blocks
}
func isPrivateIP(ipAddr string) bool {
	return isBlockAddr(ipAddr, privateBlocks)
}
func isBlockAddr(ipAddr string, blocks []*net.IPNet) bool {
	if len(blocks) == 0 {
		return true
	}
	ip := net.ParseIP(ipAddr)
	for _, priv := range blocks {

		if priv != nil && priv.Contains(ip) {
			return true
		}
	}
	return false
}
func GetLocalIps(ipBlocks []string) []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var addrs []string
	addrBlocks := genIpBlocks(ipBlocks)
	for _, iface := range ifaces {
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			// ignore error, interface can dissapear from system
			continue
		}
		for _, rawAddr := range ifaceAddrs {
			var ip net.IP
			switch addr := rawAddr.(type) {
			case *net.IPAddr:
				ip = addr.IP
			case *net.IPNet:
				ip = addr.IP
			default:
				continue
			}
			if !isBlockAddr(ip.String(), addrBlocks) {
				continue
			}
			addrs = append(addrs, ip.String())
		}

	}
	return addrs
}

// Extract returns a real ip
func Extract(addr string, ipBlocks ...string) (string, error) {
	// if addr specified then its returned
	if len(addr) > 0 && (addr != "0.0.0.0" && addr != "[::]" && addr != "::") {
		return addr, nil
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("Failed to get interfaces! Err: %v", err)
	}

	var addrs []net.Addr
	var loAddrs []net.Addr
	for _, iface := range ifaces {
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			// ignore error, interface can dissapear from system
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			loAddrs = append(loAddrs, ifaceAddrs...)
			continue
		}
		addrs = append(addrs, ifaceAddrs...)
	}
	addrs = append(addrs, loAddrs...)

	addrBlocks := genIpBlocks(ipBlocks)
	var ipAddr []byte
	var publicIP []byte

	for _, rawAddr := range addrs {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}
		if !isBlockAddr(ip.String(), addrBlocks) {
			continue
		}
		if !isPrivateIP(ip.String()) {
			publicIP = ip
			continue
		}

		ipAddr = ip
		break
	}

	// return private ip
	if ipAddr != nil {
		return net.IP(ipAddr).String(), nil
	}

	// return public or virtual ip
	if publicIP != nil {
		return net.IP(publicIP).String(), nil
	}

	return "", fmt.Errorf("No IP address found, and explicit IP not provided")
}

// IPs returns all known ips
func IPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var ipAddrs []string

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil {
				continue
			}

			// dont skip ipv6 addrs
			/*
				ip = ip.To4()
				if ip == nil {
					continue
				}
			*/

			ipAddrs = append(ipAddrs, ip.String())
		}
	}

	return ipAddrs
}
