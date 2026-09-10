package utils

import (
	"fmt"
	"testing"
)

func TestCrypto(t *testing.T) {
	str := "hwcer"
	secret := "gogogogo"
	encode, err := Crypto.DESEncrypt(str, secret)
	if err != nil {
		fmt.Printf("encode ERR:%v\n", err)
	} else {
		fmt.Printf("encode:%v\n", encode)
	}

	decode, err := Crypto.DESDecrypt(encode, secret)
	if err != nil {
		fmt.Printf("decode ERR:%v\n", err)
	} else {
		fmt.Printf("decode:%v\n", decode)
	}
}

// 随机IV语义:同一明文两次加密产生不同密文;显式IV仍由调用方管理;错误密钥返回错误而非panic
func TestCryptoRandomIV(t *testing.T) {
	original := "hello world"
	key := "1234567890123456" //AES-128

	e1, err := Crypto.AESEncrypt(original, key)
	if err != nil {
		t.Fatalf("encrypt err:%v", err)
	}
	e2, err := Crypto.AESEncrypt(original, key)
	if err != nil {
		t.Fatalf("encrypt err:%v", err)
	}
	if e1 == e2 {
		t.Error("同一明文两次加密密文相同,随机IV未生效")
	}

	d, err := Crypto.AESDecrypt(e1, key)
	if err != nil {
		t.Fatalf("decrypt err:%v", err)
	}
	if d != original {
		t.Errorf("解密结果不匹配,期望%q,实际%q", original, d)
	}

	//显式IV:密文不含IV,双方用同一IV可往返
	iv := "0123456789abcdef"
	e3, err := Crypto.AESEncrypt(original, key, iv)
	if err != nil {
		t.Fatalf("encrypt with iv err:%v", err)
	}
	d3, err := Crypto.AESDecrypt(e3, key, iv)
	if err != nil {
		t.Fatalf("decrypt with iv err:%v", err)
	}
	if d3 != original {
		t.Errorf("显式IV解密结果不匹配,期望%q,实际%q", original, d3)
	}

	//错误密钥:填充校验失败应返回错误,而不是panic或返回垃圾数据
	if _, err = Crypto.AESDecrypt(e1, "0000000000000000"); err == nil {
		t.Error("错误密钥解密应返回错误")
	}
}
