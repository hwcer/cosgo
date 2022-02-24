package request

import (
	"crypto/hmac"
	"crypto/sha1"
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

func NewOAuth(key, secret string) *OAuth {
	oauth := &OAuth{
		key:     key,
		secret:  secret,
		Strict:  true,
		Timeout: 5,
	}
	oauth.Client = New()
	oauth.Client.Header = oauth.header
	return oauth
}

type OAuth struct {
	*Client
	key     string
	secret  string
	Strict  bool  //严格模式，body会参与签名
	Timeout int32 //超时秒
}

var oauthParams = []string{"oauth_consumer_key", "oauth_nonce", "oauth_timestamp", "oauth_version", "oauth_signature_method"}

func (this *OAuth) header(method, address string, data io.Reader) map[string]string {
	header := this.NewOAuthParams()
	var body string
	if this.Strict && data != nil {
		b, _ := io.ReadAll(data)
		body = string(b)
	}
	signature := this.Signature(method, address, header, body)
	header[OAuthSignatureName] = signature
	return header
}

func (this *OAuth) NewOAuthParams() map[string]string {
	args := make(map[string]string)
	args["oauth_nonce"] = strconv.FormatInt(int64(rand.Int31n(8999)+1000), 10)
	args["oauth_version"] = "1.0"
	args["oauth_consumer_key"] = this.key
	args["oauth_signature_method"] = "HMAC-SHA1"
	args["oauth_timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	return args
}

//签名Signature
//method GET POST
//url:protocol://hostname/path
//body JSON字符串
func (this *OAuth) Signature(method, address string, oauth map[string]string, body string) string {
	arr := []string{strings.ToUpper(method), address}
	for _, k := range oauthParams {
		arr = append(arr, k+"="+oauth[k])
	}
	arr = append(arr, body)
	str := strings.Join(arr, "&")
	return HMACSHA1(this.secret, str)
}

//Verify http(s)验签
func (this *OAuth) Verify(req *http.Request) (err error) {
	signature := req.Header.Get(OAuthSignatureName)
	if signature == "" {
		return errors.New("OAuth Signature empty")
	}

	OAuthMap := make(map[string]string)
	for _, k := range oauthParams {
		v := req.Header.Get(k)
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

	var strBody string
	if this.Strict {
		var byteBody []byte
		byteBody, err = io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		strBody = string(byteBody)
	}
	if signature != this.Signature(req.Method, req.URL.String(), OAuthMap, strBody) {
		return errors.New("OAuth signature error")
	}
	return nil
}

func HMACSHA1(key, data string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
