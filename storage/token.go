package storage

import (
	crypto_rand "crypto/rand"
)

// Token 格式（28 hex 字符，14 字节原始数据）：
//
//	┌──────────┬──────────┬──────────────────────┐
//	│  bucket  │   slot   │       random         │
//	│  2 bytes │  4 bytes │       8 bytes        │
//	│  uint16  │  uint32  │  crypto/rand         │
//	└──────────┴──────────┴──────────────────────┘
//	  4 hex      8 hex        16 hex = 28 总长
//
// bucket 上限 65535，slot 上限 ~42 亿，random 提供 2^64 防猜测
// Get 路径只解析前 12 个 hex 字符（bucket+slot），剩余 16 字符通过全串比较校验

const (
	tokenBucketBytes = 2
	tokenSlotBytes   = 4
	tokenRandomBytes = 8
	tokenRawSize     = tokenBucketBytes + tokenSlotBytes + tokenRandomBytes // 14
	TokenSize        = tokenRawSize * 2                                     // 28 hex chars

	// randBufSize 随机数缓冲区大小
	// 每次 crypto/rand 系统调用获取 256 字节，可供 256/8=32 次 token 生成
	// 将 crypto/rand 开销从每次 ~100ns 摊薄到 ~3ns
	randBufSize = 256
)

const hexChars = "0123456789abcdef"

// unhexTable hex 字符 → 数值的查表，非法字符映射到 0xFF
var unhexTable [256]byte

func init() {
	for i := range unhexTable {
		unhexTable[i] = 0xFF
	}
	for i := byte('0'); i <= '9'; i++ {
		unhexTable[i] = i - '0'
	}
	for i := byte('a'); i <= 'f'; i++ {
		unhexTable[i] = i - 'a' + 10
	}
	for i := byte('A'); i <= 'F'; i++ {
		unhexTable[i] = i - 'A' + 10
	}
}

// randBuffer 随机数缓冲，减少 crypto/rand 系统调用
// 仅在 Bucket.mu.Lock 下使用，无需自身加锁
type randBuffer struct {
	buf [randBufSize]byte
	pos int
}

func (r *randBuffer) fill(dst []byte) {
	if r.pos+len(dst) > randBufSize {
		crypto_rand.Read(r.buf[:])
		r.pos = 0
	}
	copy(dst, r.buf[r.pos:r.pos+len(dst)])
	r.pos += len(dst)
}

// tokenEncode 生成 token 字符串
// bucket 和 slot 编码为固定宽度 hex，random 部分从 randBuffer 获取
// 整个过程仅 1 次堆分配（string 转换）
func tokenEncode(bucket uint16, slot uint32, rng *randBuffer) string {
	var buf [TokenSize]byte

	// bucket: 2 bytes → 4 hex chars
	buf[0] = hexChars[bucket>>12&0xF]
	buf[1] = hexChars[bucket>>8&0xF]
	buf[2] = hexChars[bucket>>4&0xF]
	buf[3] = hexChars[bucket&0xF]

	// slot: 4 bytes → 8 hex chars
	buf[4] = hexChars[slot>>28&0xF]
	buf[5] = hexChars[slot>>24&0xF]
	buf[6] = hexChars[slot>>20&0xF]
	buf[7] = hexChars[slot>>16&0xF]
	buf[8] = hexChars[slot>>12&0xF]
	buf[9] = hexChars[slot>>8&0xF]
	buf[10] = hexChars[slot>>4&0xF]
	buf[11] = hexChars[slot&0xF]

	// random: 8 bytes → 16 hex chars
	var raw [tokenRandomBytes]byte
	rng.fill(raw[:])
	for i, b := range raw {
		buf[12+i*2] = hexChars[b>>4]
		buf[12+i*2+1] = hexChars[b&0xF]
	}

	return string(buf[:])
}

// tokenDecodeBucket 从 token 前 4 个 hex 字符解析 bucket 索引
// 零分配，纯查表运算，~2ns
func tokenDecodeBucket(token string) (int, bool) {
	if len(token) != TokenSize {
		return 0, false
	}
	h0 := unhexTable[token[0]]
	h1 := unhexTable[token[1]]
	h2 := unhexTable[token[2]]
	h3 := unhexTable[token[3]]
	if h0|h1|h2|h3 == 0xFF {
		return 0, false
	}
	return int(h0)<<12 | int(h1)<<8 | int(h2)<<4 | int(h3), true
}

// tokenDecodeSlot 从 token 第 4-11 个 hex 字符解析 slot 索引
// 零分配，纯查表运算，~3ns
func tokenDecodeSlot(token string) (int, bool) {
	if len(token) != TokenSize {
		return 0, false
	}
	var val uint32
	for i := 4; i < 12; i++ {
		h := unhexTable[token[i]]
		if h == 0xFF {
			return 0, false
		}
		val = val<<4 | uint32(h)
	}
	return int(val), true
}
