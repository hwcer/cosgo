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
	Port      int    `json:"port"`
	Host      string `json:"host"`
	Retry     int    `json:"retry"`
	Scheme    string `json:"scheme"`
	localIpv4 string
}

func (this *Address) Parse(address string) {
	if strings.Contains(address, "://") {
		arr := strings.Split(address, "://")
		address = arr[1]
		this.Scheme = strings.ToLower(arr[0])
	}
	pair := strings.Split(address, ":")
	this.Host = pair[0]
	if len(pair) > 1 {
		this.Port, _ = strconv.Atoi(pair[1])
	}
	return
}

// String 转换成string
func (this *Address) String(withScheme ...bool) string {
	b := strings.Builder{}
	if len(withScheme) > 0 && withScheme[0] && this.Scheme != "" {
		b.WriteString(this.Scheme)
		b.WriteString("://")
	}
	if this.Host != "" {
		b.WriteString(this.Host)
	}
	if this.Port > 0 {
		b.WriteString(":")
		b.WriteString(strconv.Itoa(this.Port))
	}
	return b.String()
}
func (this *Address) Empty() bool {
	return this.Host == "" || this.Host == "0.0.0.0" || this.Host == "127.0.0.1" || this.Host == "localhost"
}
func (this *Address) LocalIPv4() string {
	if this.localIpv4 == "" {
		var err error
		if this.localIpv4, err = LocalIpv4(); err != nil {
			this.localIpv4 = "0.0.0.0"
		}
	}
	return this.localIpv4
}

func (this *Address) Encode() uint64 {
	addr := this.String(false)
	return Ipv4Encode(addr)

}

func (this *Address) Handle(handle func(network, address string) error) (err error) {
	address := this.String(false)
	if err = handle(this.Scheme, address); err == nil || !IsOsBindError(err) {
		return
	}
	this.Port += 1
	this.Retry -= 1
	if this.Retry <= 0 {
		return
	}
	return this.Handle(handle)
}

// HandleWithNetwork network 写入地址中,tcp://0.0.0.0:80
func (this *Address) HandleWithNetwork(handle func(address string) error) (err error) {
	address := this.String(true)
	if err = handle(address); err == nil || !IsOsBindError(err) {
		return
	}
	this.Port += 1
	this.Retry -= 1
	if this.Retry <= 0 {
		return
	}
	return this.HandleWithNetwork(handle)
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

func LocalIpv4() (ip string, err error) {
	var ipv4 []string
	if ipv4, err = LocalIPv4s(); err != nil {
		return
	}
	for _, s := range ipv4 {
		i := strings.Index(s, ".")
		if k := s[:i]; k == "192" || k == "10" || k == "172" {
			return s, nil
		}
	}
	if len(ipv4) == 0 {
		err = fmt.Errorf("无法获取服务器的内网IP")
	} else {
		ip = ipv4[0]
	}
	return
}

// LocalIPv4s return all non-loopback IPv4 addresses
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

// Ipv4Encode  Ip2Int Ipv4 转uint64
func Ipv4Encode(address string) uint64 {
	var ip string
	var port string
	i := strings.Index(address, ":")
	if i > 0 {
		ip = address[0:i]
		port = address[i+1:]
	} else {
		ip = address
	}

	ip = strings.TrimSpace(ip)
	ips := strings.Split(ip, ".")
	var r uint64 = 0
	var pos uint64 = 24
	for _, ipSeg := range ips {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		r = r | uint64(tempInt)
		pos -= 8
	}
	if port != "" {
		p, _ := strconv.Atoi(port)
		r += uint64(p << 32)
	}
	return r
}

func Ipv4Decode(code uint64) string {
	ip := uint32(code & 0xffffffff)
	ips := make([]string, 4)
	ips[0] = fmt.Sprintf("%v", ip>>24)
	ips[1] = fmt.Sprintf("%v", (ip&0x00ff0000)>>16)
	ips[2] = fmt.Sprintf("%v", (ip&0x0000ff00)>>8)
	ips[3] = fmt.Sprintf("%v", ip&0x000000ff)
	var arr []string
	arr = append(arr, strings.Join(ips, "."))
	port := code >> 32
	arr = append(arr, strconv.Itoa(int(port)))
	return strings.Join(arr, ":")
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
