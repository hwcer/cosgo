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
