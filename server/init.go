package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

var Config = struct {
	ErrHeaderName string //将错误码写入HTTP头报文名字

}{
	ErrHeaderName: "X-Error-Code",
}

var OAuth  *oauth

func init() {
	OAuth = NewOAuth()
}

func HMACSHA1(key, data string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

