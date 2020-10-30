package express

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

/*
	基于GIN的 OAUTH1.0 签名与身份认证
	URL中的query参数一律不参与签名
*/

type oauth struct {
	Key     string
	Secret  string
	Strict  bool  //严格模式，body会参与签名
	Timeout int32 //超时秒
}

const OAuth_Signature_Name = "oauth_signature"

var oauthParams = []string{"oauth_consumer_key", "oauth_nonce", "oauth_timestamp", "oauth_version", "oauth_signature_method"}

func NewOAuth() *oauth {
	return &oauth{Key: "oauth1.0", Secret: "szmzbzbzlp@20200712", Timeout: 5}
}

func (this *oauth) NewOAuthParams() map[string]string {
	oauth := make(map[string]string)
	oauth["oauth_consumer_key"] = this.Key
	oauth["oauth_version"] = "1.0"
	oauth["oauth_signature_method"] = "HMAC-SHA1"
	oauth["oauth_nonce"] = strconv.FormatInt(int64(rand.Int31n(8999)+1000), 10)
	oauth["oauth_timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	return oauth
}
func HMACSHA1(key, data string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

//签名Signature
//method GET Insert
//url:protocol://hostname/path
//body JSON字符串
func (this *oauth) Signature(method, url string, oauth map[string]string, body string) string {
	arr := []string{method, url}
	for _, k := range oauthParams {
		arr = append(arr, k+"="+oauth[k])
	}
	arr = append(arr, body, this.Secret)
	str := strings.Join(arr, "&")
	return HMACSHA1(this.Secret, str)
}

//Verify http(s)验签
func (this *oauth) Verify(ctx *gin.Context) error {
	signature := ctx.GetHeader(OAuth_Signature_Name)
	if signature == "" {
		return errors.New("OAuth Signature empty")
	}

	OAuthMap := make(map[string]string)
	for _, k := range oauthParams {
		v := ctx.GetHeader(k)
		if v == "" {
			return errors.New("OAuth Params empty")
		}
		OAuthMap[k] = v
	}
	if OAuthMap["oauth_consumer_key"] != this.Key {
		return errors.New("oauth_consumer_key error")
	}
	oauthTimeStamp, err := strconv.ParseInt(OAuthMap["oauth_timestamp"], 10, 64)
	if err != nil {
		return err
	}

	requestTime := time.Now().Unix() - oauthTimeStamp
	if requestTime < 0 || requestTime > int64(this.Timeout) {
		return errors.New("oauth request timeout")
	}

	var strBody string
	if this.Strict {
		byteBody, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			return err
		}
		strBody = string(byteBody)
	}

	baseUrl := strings.Join([]string{ctx.Request.URL.Scheme, "://", ctx.Request.URL.Host, ctx.Request.URL.EscapedPath()}, "")
	newSignature := this.Signature(ctx.Request.Method, baseUrl, OAuthMap, strBody)

	if signature != newSignature {
		return errors.New("oauth signature error")
	}
	return nil
}

//args
func (this *oauth) request(method, rawurl string, data []byte, header map[string]string) *Message {
	var (
		err   error
		req   *http.Request
		res   *http.Response
		reply []byte
	)
	//添加随机参数
	var urlPase *url.URL
	urlPase, err = url.Parse(rawurl)
	if err != nil {
		return NewErrMsgFromError(err)
	}
	query := urlPase.Query()
	query.Set("_", strconv.Itoa(int(time.Now().Unix())))
	urlPase.RawQuery = query.Encode()

	req, err = http.NewRequest(method, urlPase.String(), bytes.NewBuffer(data))
	if err != nil {
		return NewErrMsgFromError(err)
	}
	defer req.Body.Close()

	if header == nil {
		header = make(map[string]string)
	}

	if _, ok := header["content-type"]; !ok {
		header["content-type"] = "application/json;charset=utf-8"
	}
	for k, v := range header {
		req.Header.Add(k, v)
	}
	//设置签名
	OAuthMap := this.NewOAuthParams()
	for k, v := range OAuthMap {
		req.Header.Add(k, v)
	}

	var strBody string
	if this.Strict {
		strBody = string(data)
	}
	baseUrl := strings.Join([]string{urlPase.Scheme, "://", urlPase.Host, urlPase.EscapedPath()}, "")
	signature := this.Signature(method, baseUrl, OAuthMap, strBody)
	req.Header.Add(OAuth_Signature_Name, signature)

	client := &http.Client{Timeout: time.Duration(this.Timeout) * time.Second}
	res, err = client.Do(req)
	if err != nil {
		return NewErrMsgFromError(err)
	}
	defer res.Body.Close()

	reply, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return NewErrMsgFromError(err)
	} else if res.StatusCode != http.StatusOK {
		return NewErrMsg(res.Status)
	}

	if Config.ErrHeaderName != "" {
		var code int
		code, err = strconv.Atoi(res.Header.Get(Config.ErrHeaderName))
		if err != nil {
			return NewErrMsgFromError(err)
		} else if code != GetErrCode(ErrMsg_NAME_SUCCESS) {
			return NewErrMsg(string(reply), code)
		}
	}
	return NewMsg(reply)
}

func (this *oauth) GET(url string, data []byte, header map[string]string) *Message {
	return this.request("GET", url, data, header)
}

func (this *oauth) POST(url string, data []byte, header map[string]string) *Message {
	return this.request("POST", url, data, header)
}

func (this *oauth) PostJson(url string, query interface{}) *Message {
	data, err := json.Marshal(query)
	if err != nil {
		return NewErrMsgFromError(err)
	}
	return this.request("POST", url, data, nil)
}
