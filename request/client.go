package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hwcer/cosgo/binder"
)

// httpClient 默认 HTTP 客户端，替代 http.DefaultClient
// 配置合理的连接池和超时，防止生产环境踩坑：
//   - MaxIdleConnsPerHost: http.DefaultClient 默认只有 2，高并发对同一 host 会频繁创建/销毁连接
//   - Timeout: http.DefaultClient 默认无超时，目标服务器无响应时 goroutine 永久挂起
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 32,
		IdleConnTimeout:     90 * time.Second,
	},
}

func New() *Client {
	r := &Client{}
	r.Binder = binder.New(binder.MIMEJSON)
	return r
}

type middleware func(req *http.Request) error

// Client HTTP 客户端
// 支持中间件链、OAuth 签名、自动 Content-Type 协商
type Client struct {
	oauth      *OAuth
	Binder     binder.Binder
	middleware []middleware
}

// Use 注册请求前中间件（按注册顺序执行）
func (c *Client) Use(m middleware) {
	c.middleware = append(c.middleware, m)
}

// OAuth 初始化 OAuth1 签名器
func (c *Client) OAuth(key, secret string, strict ...bool) *OAuth {
	c.oauth = NewOAuth(key, secret, strict...)
	return c.oauth
}

// Verify 服务端验证请求的 OAuth 签名
func (c *Client) Verify(req *http.Request, body *bytes.Buffer) (err error) {
	if c.oauth == nil {
		return nil
	}
	return c.oauth.Verify(req, body)
}

// Marshal 将数据序列化为 bytes.Buffer
// 优先处理 []byte / string / io.Reader 零拷贝路径，其余走 Binder 编码
func (c *Client) Marshal(data any) (b *bytes.Buffer, err error) {
	b = bytes.NewBuffer(nil)
	if data == nil {
		return
	}
	switch v := data.(type) {
	case []byte:
		b.Write(v)
	case *[]byte:
		b.Write(*v)
	case string:
		b.WriteString(v)
	case *string:
		b.WriteString(*v)
	case io.Reader:
		_, err = b.ReadFrom(v)
	default:
		err = c.Binder.Encode(b, data)
	}
	return
}

// Request 执行 HTTP 请求
// 流程：Marshal → NewRequest → 设置 Header → 中间件链 → OAuth 签名 → 发送 → 读响应
func (c *Client) Request(method, url string, data any, header ...map[string]string) (reply []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	var b *bytes.Buffer
	if b, err = c.Marshal(data); err != nil {
		return
	}
	var req *http.Request
	if req, err = http.NewRequest(method, url, b); err != nil {
		return
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	if len(header) > 0 {
		for k, v := range header[0] {
			req.Header.Set(k, v)
		}
	}
	if ct := req.Header.Get("Content-Type"); ct == "" {
		if contentType := c.Binder.String(); contentType != "" {
			req.Header.Add("Content-Type", FormatContentTypeAndCharset(contentType))
		}
	}

	for _, m := range c.middleware {
		if err = m(req); err != nil {
			return
		}
	}
	if c.oauth != nil {
		if err = c.oauth.Request(req); err != nil {
			return
		}
	}

	var res *http.Response
	res, err = httpClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	reply, err = io.ReadAll(res.Body)
	return
}

// Get 发送 GET 请求并自动解码响应
// reply 类型：*[]byte 原始字节、*string 字符串、其他类型走 Binder 反序列化
func (c *Client) Get(url string, reply interface{}) (err error) {
	var body []byte
	body, err = c.Request(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	switch v := reply.(type) {
	case *[]byte:
		*v = body
	case *string:
		*v = string(body)
	default:
		err = c.Binder.Unmarshal(body, reply)
	}
	return
}

// Post 发送 POST 请求并自动解码响应
func (c *Client) Post(url string, data interface{}, reply interface{}) (err error) {
	var body []byte
	body, err = c.Request(http.MethodPost, url, data)
	if err != nil || reply == nil {
		return
	}
	switch v := reply.(type) {
	case *[]byte:
		*v = body
	case *string:
		*v = string(body)
	default:
		err = c.Binder.Unmarshal(body, reply)
	}
	return
}
