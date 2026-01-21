// Package session 提供会话管理功能，支持内存和Redis存储
package session

const ContextRandomStringLength = 6

// Options 全局配置选项
// 注意：
// 1. Name: session cookie name
// 2. MaxAge: 会话有效期，单位为秒
// 3. Secret: 16位秘钥，用于Redis存储时生成TOKEN
// 4. Storage: 存储后端，支持内存和Redis存储
// 5. Heartbeat: 心跳间隔，单位为秒，用于自动清理过期会话
var Options = struct {
	Name string //session cookie name
	//Token     token  //token生成和解析方式
	MaxAge    int64  //有效期(S)
	Secret    string //16位秘钥
	Storage   Storage
	Heartbeat int32 //心跳(S)

}{
	Name:      "_cookie_vars",
	MaxAge:    3600,
	Secret:    "UVFGHIJABCopqDNO", //redis 存储时生成TOKEN的密钥
	Heartbeat: 10,
}

//func Decode(sid string) (uid string, err error) {
//	if Options.Token != nil {
//		return Options.Token.Decode(sid)
//	}
//	str, err := utils.Crypto.AESDecrypt(sid, Options.Secret)
//	if err != nil {
//		return "", err
//	}
//	uid = str[ContextRandomStringLength:]
//	return
//}
//
//func Encode(uid string) (sid string, err error) {
//	if Options.Token != nil {
//		return Options.Token.Encode(uid)
//	}
//	var arr []string
//	arr = append(arr, random.Strings.String(ContextRandomStringLength))
//	arr = append(arr, uid)
//	str := strings.Join(arr, "")
//	return utils.Crypto.AESEncrypt(str, Options.Secret)
//}
