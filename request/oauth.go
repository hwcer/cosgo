package request

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
	OAUTH1.0 签名与身份认证
*/

const OAuthSignatureName = "oauth_signature"
const (
	HeaderXForwardedProto    = "X-Forwarded-Code"
	HeaderXForwardedProtocol = "X-Forwarded-Protocol"
	HeaderXForwardedSsl      = "X-Forwarded-Ssl"
	HeaderXUrlScheme         = "X-Url-Protocol"
)

func NewOAuth(key, secret string, strict ...bool) *OAuth {
	oauth := &OAuth{key: key, secret: secret, Strict: false, Timeout: 30}
	if len(strict) > 0 {
		oauth.Strict = strict[0]
	}
	return oauth
}

type Header interface {
	Get(key string) string
}

type OAuth struct {
	key     string
	secret  string
	Strict  bool  //严格模式，body会参与签名
	Timeout int32 //超时秒
}

var oauthParams = []string{"oauth_consumer_key", "oauth_nonce", "oauth_timestamp", "oauth_version", "oauth_signature_method"}

func (this *OAuth) NewOAuthParams() map[string]string {
	args := make(map[string]string)
	args["oauth_nonce"] = strconv.FormatInt(int64(rand.Int31n(8999)+1000), 10)
	args["oauth_version"] = "1.0"
	args["oauth_consumer_key"] = this.key
	args["oauth_signature_method"] = "HMAC-SHA1"
	args["oauth_timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	return args
}

// Signature 签名Signature
// method GET POST
// body JSON字符串
//
//url:protocol://hostname/path
func (this *OAuth) Signature(address string, oauth map[string]string, body string) string {
	arr := []string{address}
	for _, k := range oauthParams {
		arr = append(arr, k+"="+oauth[k])
	}
	arr = append(arr, body)
	str := strings.Join(arr, "&")
	return HMACSHA256(this.secret, str)
}

// Request 自动设置HTTP请求头
func (this *OAuth) Request(req *http.Request) (err error) {
	header := this.NewOAuthParams()
	var bodyBytes []byte
	if this.Strict {
		var body io.ReadCloser
		if body, err = req.GetBody(); err != nil {
			return
		}
		if bodyBytes, err = io.ReadAll(body); err != nil {
			return
		}
	}
	signature := this.Signature(Address(req), header, string(bodyBytes))
	header[OAuthSignatureName] = signature
	for k, v := range header {
		req.Header.Add(k, v)
	}
	return nil
}

// Verify http(s)验签
func (this *OAuth) Verify(req *http.Request, body *bytes.Buffer) (err error) {
	header := req.Header
	address := Address(req)
	signature := header.Get(OAuthSignatureName)
	if signature == "" {
		return errors.New("OAuth Signature empty")
	}

	OAuthMap := make(map[string]string)
	for _, k := range oauthParams {
		v := header.Get(k)
		if v == "" {
			return errors.New("OAuth Params empty")
		}
		OAuthMap[k] = v
	}
	if OAuthMap["oauth_consumer_key"] != this.key {
		return errors.New("oauth_consumer_key error")
	}

	//验证时间
	if this.Timeout > 0 {
		var oauthTimeStamp int64
		oauthTimeStamp, err = strconv.ParseInt(OAuthMap["oauth_timestamp"], 10, 64)
		if err != nil {
			return err
		}
		requestTime := time.Now().Unix() - oauthTimeStamp
		if requestTime < 0 || requestTime > int64(this.Timeout) {
			return errors.New("OAuth timeout")
		}
	}

	var data string
	if this.Strict {
		if body == nil {
			body = bytes.NewBuffer(nil)
			if _, err = body.ReadFrom(req.Body); err != nil {
				return err
			}
			req.Body = io.NopCloser(body)
		}
		data = body.String()
	}
	if signature != this.Signature(address, OAuthMap, data) {
		return errors.New("OAuth signature error")
	}
	return
}

func HMACSHA1(key, data string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
func HMACSHA256(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
