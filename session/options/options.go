package options

import (
	"github.com/hwcer/cosgo/utils"
	"strings"
)

const SessionContextRandomStringLength = 4

var Options = struct {
	Name   string //session cookie name
	MaxAge int64  //有效期(S)
	Secret string //16位秘钥
}{
	Name:   "_cosweb_cookie_vars",
	MaxAge: 3600,
	Secret: "UVFGHIJABCopqDNO",
}

func Decode(sid string) (key string, err error) {
	str, err := utils.Crypto.AESDecrypt(sid, Options.Secret)
	if err != nil {
		return "", err
	}
	key = str[SessionContextRandomStringLength:]
	return
}

func Encode(key string) (sid string, err error) {
	var arr []string
	arr = append(arr, utils.Random.String(SessionContextRandomStringLength))
	arr = append(arr, key)
	str := strings.Join(arr, "")
	return utils.Crypto.AESEncrypt(str, Options.Secret)
}
