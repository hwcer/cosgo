package session

import (
	"github.com/hwcer/cosgo/random"
	"github.com/hwcer/cosgo/utils"
	"strings"
)

const ContextRandomStringLength = 4

type token interface {
	Decode(sid string) (uid string, err error)
	Encode(uid string) (sid string, err error)
}

var Options = struct {
	Name    string //session cookie name
	Token   token  //token生成和解析方式
	MaxAge  int64  //有效期(S)
	Secret  string //16位秘钥
	Storage Storage
}{
	Name:   "_cosweb_cookie_vars",
	MaxAge: 3600,
	Secret: "UVFGHIJABCopqDNO",
}

func Decode(sid string) (uid string, err error) {
	if Options.Token != nil {
		return Options.Token.Decode(sid)
	}
	str, err := utils.Crypto.AESDecrypt(sid, Options.Secret)
	if err != nil {
		return "", err
	}
	uid = str[ContextRandomStringLength:]
	return
}

func Encode(uid string) (sid string, err error) {
	if Options.Token != nil {
		return Options.Token.Encode(uid)
	}
	var arr []string
	arr = append(arr, random.Strings.String(ContextRandomStringLength))
	arr = append(arr, uid)
	str := strings.Join(arr, "")
	return utils.Crypto.AESEncrypt(str, Options.Secret)
}
