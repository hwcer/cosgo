package safety

import "strconv"

// parseIPv4 将 IPv4 字符串解析为 uint32，零堆分配
// 自动剥离端口号（如 "192.168.1.2:8080" → 只取 IP 部分）
// 非法输入返回 0
func parseIPv4(ip string) uint32 {
	// 剥离端口
	for i := 0; i < len(ip); i++ {
		if ip[i] == ':' {
			ip = ip[:i]
			break
		}
	}

	var result uint32
	var octet uint32
	var dots int

	for i := 0; i < len(ip); i++ {
		c := ip[i]
		switch {
		case c >= '0' && c <= '9':
			octet = octet*10 + uint32(c-'0')
		case c == '.':
			if octet > 255 {
				return 0
			}
			result = result<<8 | octet
			octet = 0
			dots++
		case c == ' ' || c == '\t':
			continue
		default:
			return 0
		}
	}
	if dots != 3 || octet > 255 {
		return 0
	}
	return result<<8 | octet
}

// parseRule 解析规则字符串为 IP 范围 [start, end]
// 支持三种格式：
//   - 单 IP：    "10.0.0.1"               → start=end
//   - 范围：     "10.0.0.0~10.255.255.255" → [start, end]
//   - CIDR：    "10.0.0.0/8"              → 自动计算范围
func parseRule(rule string) (start, end uint32) {
	// 检测 CIDR
	for i := 0; i < len(rule); i++ {
		if rule[i] == '/' {
			return parseCIDR(rule[:i], rule[i+1:])
		}
	}
	// 检测范围
	for i := 0; i < len(rule); i++ {
		if rule[i] == '~' {
			return parseIPv4(rule[:i]), parseIPv4(rule[i+1:])
		}
	}
	// 单 IP
	ip := parseIPv4(rule)
	return ip, ip
}

// parseCIDR 将 CIDR 表示法转为 IP 范围
// 例如 "10.0.0.0", "8" → start=10.0.0.0, end=10.255.255.255
func parseCIDR(ipStr, bitsStr string) (start, end uint32) {
	ip := parseIPv4(ipStr)
	bits, err := strconv.Atoi(bitsStr)
	if err != nil || bits < 0 || bits > 32 {
		return 0, 0
	}
	if bits == 0 {
		return 0, 0xFFFFFFFF
	}
	mask := uint32(0xFFFFFFFF) << (32 - bits)
	return ip & mask, ip | ^mask
}
