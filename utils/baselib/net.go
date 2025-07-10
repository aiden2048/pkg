package baselib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type INetAddr struct {
	Host string
	Port int
}

var localIp string = ""

func GetLocalIP() string {
	if localIp != "" {
		return localIp
	}

	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIp = ipnet.IP.String()
				break
			}
		}
	}
	return localIp
}

var u32LocalIP uint32 = 0

// GetLocalIP return local eth1 ip
func GetU32LocalIP() uint32 {
	if u32LocalIP != 0 {
		return u32LocalIP
	}
	u32LocalIP = IPtoI(GetIP("eth1"))

	return u32LocalIP
}

func getIPFromInterface(v net.Interface) string {
	addrs, err := v.Addrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		var ip net.IP
		switch val := addr.(type) {
		case *net.IPNet:
			ip = val.IP
		case *net.IPAddr:
			ip = val.IP
		}
		return ip.String()
	}

	return ""
}

// 获取网卡的IP，如eth0、eth1
// 传入多个时，按顺序查找是否存在，返回找到的第一个有效网卡的IP
func GetIP(names ...string) string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, name := range names {
		for _, v := range ifaces {
			if v.Name == name {
				if ip := getIPFromInterface(v); ip != "" {
					return ip
				}
			}
		}
	}
	return ""
}

// 获取 字典序最小的网卡 的IP
func GetFirstNameIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	itf := net.Interface{Name: ""}
	ip := ""
	for _, inter := range ifaces {
		// 剔除lo网卡
		if (inter.Flags&net.FlagLoopback) == 0 && (itf.Name == "" || strings.Compare(itf.Name, inter.Name) > 0) {
			if tmpIP := getIPFromInterface(inter); tmpIP != "" {
				itf, ip = inter, tmpIP
			}
		}
	}

	return ip
}

func GetFirstMacAddr() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	itf := net.Interface{Name: ""}
	for _, inter := range interfaces {
		// fmt.Println(inter)
		if (inter.Flags&net.FlagLoopback) == 0 && inter.Name != "eth0" && (itf.Name == "" || strings.Compare(itf.Name, inter.Name) > 0) {
			itf = inter
		}
	}
	return itf.HardwareAddr.String(), nil
}

// IP格式转为uint32
func IP2Uint32(ipStr string) uint32 {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

// uint32转为IP格式
func Uint32ToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip>>24, ip<<8>>24, ip<<16>>24, ip<<24>>24)
}

// uint64转为IP字符串
func Uint64ToIP(ip uint64) string {
	result := make(net.IP, 4)
	result[3] = byte(ip)
	result[2] = byte(ip >> 8)
	result[1] = byte(ip >> 16)
	result[0] = byte(ip >> 24)
	return result.String()
}

// IPtoI convert ip from string to uint32, like 10.100.67.132 to 174343044, if fail return 0
func IPtoI(ip string) uint32 {
	ips := net.ParseIP(ip)

	if len(ips) == 16 {
		return binary.BigEndian.Uint32(ips[12:16])
	} else if len(ips) == 4 {
		return binary.BigEndian.Uint32(ips)
	}
	return 0
}

// ItoIP convert ip from uint32 to string, like 174343044 to 10.100.67.132, if fail return empty string: ""
func ItoIP(ip uint32) string {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, ip)
	if err != nil {
		return ""
	}

	b := buf.Bytes()
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}

const X_RemoteAddr = "X-Addr"

func FixupRealIpForHttp(r *http.Request) {
	if r == nil {
		return
	}

	cdnIp := r.Header["Cf-Connecting-Ip"]
	if len(cdnIp) > 0 {
		ipListStr := strings.Replace(cdnIp[0], " ", "", -1)
		ipList := strings.Split(ipListStr, ",")
		ip0 := net.ParseIP(ipList[0])
		if len(ipList) > 0 && ip0 != nil && !ip0.IsLoopback() {
			r.Header[X_RemoteAddr] = []string{r.RemoteAddr}
			r.RemoteAddr = ipList[0]

			return
		}
	}
	xRealIp := r.Header["X-Forwarded-For"]
	if len(xRealIp) > 0 {
		ipListStr := strings.Replace(r.Header["X-Forwarded-For"][0], " ", "", -1) // 这里拿到的是 127.0.0.1,120.0.0.2
		ipList := strings.Split(ipListStr, ",")
		if len(ipList) <= 0 {
			return
		}
		remoteIP := ipList[0]
		if strings.Count(remoteIP, ":") == 1 { // ipv4
			remoteIP = strings.Split(remoteIP, ":")[0]
		}
		ip0 := net.ParseIP(remoteIP)
		if len(ipList) > 0 && ip0 != nil && !ip0.IsLoopback() {
			r.Header[X_RemoteAddr] = []string{r.RemoteAddr}
			r.RemoteAddr = ip0.String()
			// fmt.Println("FixupRealIpForHttp r.RemoteAddr:%+v", r.RemoteAddr)
			return
		}

	}
	r.RemoteAddr = GetIpFromAddress(r.RemoteAddr)
}

func GetIpFromAddress(addr string) string {
	if strings.Count(addr, ":") == 1 {
		n := strings.Index(addr, ":")
		addr = addr[0:n]
	}
	return addr
}

func GetAddresFromHttpHeader(header map[string][]string, addr string) string {
	if header == nil {
		return addr
	}
	addr = GetIpFromAddress(addr)
	xRealIp := header["X-Forwarded-For"]
	if len(xRealIp) > 0 {
		ipListStr := strings.Replace(header["X-Forwarded-For"][0], " ", "", -1) // 这里拿到的是 127.0.0.1,120.0.0.2
		ipList := strings.Split(ipListStr, ",")
		if len(ipList) <= 0 {
			return addr
		}
		remoteIP := ipList[0]
		if strings.Count(remoteIP, ":") == 1 { // ipv4
			remoteIP = strings.Split(remoteIP, ":")[0]
		}
		ip0 := net.ParseIP(remoteIP)
		if len(ipList) > 0 && ip0 != nil && !ip0.IsLoopback() {
			addr = remoteIP
		}
	}
	return addr
}
