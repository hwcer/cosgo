package random

import (
	cryptorand "crypto/rand"
)

// Strings 默认的随机字符串生成器，使用 crypto/rand 作为熵源，适合生成 token/secret。
// 注意：游戏概率类随机请用 Roll / Probability / Relative（math/rand）。
var Strings = NewBytes([]byte("0123456789abcdefghijklmnopqrstuvwxyz"))

// Bytes 基于自定义字符集 + crypto/rand 的安全随机字符串生成器。
type Bytes struct {
	bytes []byte
	limit int // 拒绝采样阈值，预计算避免每次 New 重算
}

func NewBytes(bytes []byte) *Bytes {
	b := &Bytes{bytes: bytes}
	if n := len(bytes); n >= 2 {
		b.limit = 256 - (256%n)
	}
	return b
}

// New 返回长度为 l 的随机字节串，每个字节来自字符集。
// 使用 crypto/rand + 拒绝采样消除模偏差。
// 单次分配，原地处理：读 l 字节后就地筛选，极少情况需要补读。
func (this *Bytes) New(l int) []byte {
	if l <= 0 {
		return nil
	}
	n := len(this.bytes)
	if n == 0 {
		return make([]byte, l)
	}
	if n == 1 {
		r := make([]byte, l)
		for i := range r {
			r[i] = this.bytes[0]
		}
		return r
	}

	// 单次分配：读入 buf，原地将 crypto 字节映射为字符集字节
	buf := make([]byte, l)
	cryptorand.Read(buf)

	limit := this.limit
	w := 0
	for _, b := range buf {
		if int(b) < limit {
			buf[w] = this.bytes[int(b)%n]
			w++
		}
	}

	// 典型字符集（36 字符）拒绝率仅 ~1.6%，几乎不会进入此循环
	for w < l {
		var extra [32]byte
		cryptorand.Read(extra[:])
		for _, b := range extra {
			if w >= l {
				break
			}
			if int(b) < limit {
				buf[w] = this.bytes[int(b)%n]
				w++
			}
		}
	}

	return buf[:l]
}

func (this *Bytes) String(l int) string {
	return string(this.New(l))
}
