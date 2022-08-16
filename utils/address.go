package utils

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Address struct {
	Port   int    `json:"port"`
	Host   string `json:"host"`
	Retry  int    `json:"retry"`
	Scheme string `json:"scheme"`
}

func (this *Address) Parse(address string) {
	if strings.Contains(address, "://") {
		arr := strings.Split(address, "://")
		address = arr[1]
		this.Scheme = strings.ToLower(arr[0])
	}
	pair := strings.Split(address, ":")
	this.Host = pair[0]
	if len(pair) > 0 {
		this.Port, _ = strconv.Atoi(pair[1])
	}
	return
}

// String 转换成string   scheme : tcp  OR   tcp,://
func (this *Address) String(scheme ...string) string {
	b := strings.Builder{}
	if len(scheme) > 0 {
		for _, k := range scheme {
			b.WriteString(k)
		}
	} else if this.Scheme != "" {
		b.WriteString(this.Scheme)
		b.WriteString("://")
	}
	if this.Host != "" {
		b.WriteString(this.Host)
	}
	b.WriteString(":")
	b.WriteString(strconv.Itoa(this.Port))
	return b.String()
}

func (this *Address) Handle(handle func(string) error) (err error) {
	address := this.String()
	if err = handle(address); err == nil || !IsOsBindError(err) {
		return
	}
	this.Port += 1
	this.Retry -= 1
	if this.Retry <= 0 {
		return
	}
	return this.Handle(handle)
}

// NewAddress 解析url,scheme:默认协议
func NewAddress(address ...string) (r *Address) {
	r = &Address{}
	if len(address) > 0 {
		r.Parse(address[0])
	}
	return
}

func NewUrl(address, scheme string) (*url.URL, error) {
	if !strings.Contains(address, "://") {
		address = scheme + "://" + address
	}
	return url.Parse(address)
}

// LocalIPs return all non-loopback IPv4 addresses
func LocalIPv4s() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}
	for _, a := range addrs {
		if ip, ok := isLocalIpv4(a); ok {
			ips = append(ips, ip)
		}
	}

	return ips, nil
}

func isLocalIpv4(a net.Addr) (string, bool) {
	if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
		ip := ipnet.IP.String()
		if strings.HasSuffix(ip, ".1") || strings.HasSuffix(ip, ".255") {
			return "", false
		} else {
			return ip, true
		}
	}
	return "", false
}

// Ip2Int Ipv4 转uint32
func Ipv4Encode(ip string) uint32 {
	if i := strings.Index(ip, ":"); i > 0 {
		ip = ip[0:i]
	}
	ip = strings.TrimSpace(ip)
	ips := strings.Split(ip, ".")
	var ipCode int = 0
	var pos uint = 24
	for _, ipSeg := range ips {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipCode = ipCode | tempInt
		pos -= 8
	}
	return uint32(ipCode)
}

func Ipv4Decode(ipCode uint32) string {
	ips := make([]string, 4)
	ips[0] = fmt.Sprintf("%v", ipCode>>24)
	ips[1] = fmt.Sprintf("%v", (ipCode&0x00ff0000)>>16)
	ips[2] = fmt.Sprintf("%v", (ipCode&0x0000ff00)>>8)
	ips[3] = fmt.Sprintf("%v", ipCode&0x000000ff)
	return strings.Join(ips, ".")
}

// GetIPv4ByInterface return IPv4 address from a specific interface IPv4 addresses
func GetIPv4ByInterface(name string) ([]string, error) {
	var ips []string

	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}

// IsOsBindError 是否端口绑定错误
func IsOsBindError(err error) bool {
	var ok bool
	var opErr *net.OpError
	if opErr, ok = err.(*net.OpError); !ok {
		return false
	}
	var syscallErr *os.SyscallError
	if syscallErr, ok = opErr.Err.(*os.SyscallError); !ok {
		return false
	}
	return syscallErr.Syscall == "bind"
}
