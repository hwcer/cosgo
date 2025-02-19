package request

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/binder"
	"io"
	"net/http"
)

func New() *Client {
	r := &Client{}
	r.Binder = binder.New(binder.MIMEJSON)
	return r
}

type middleware func(req *http.Request) error

type Client struct {
	oauth      *OAuth
	Binder     binder.Binder
	middleware []middleware
}

// Use 发放请求前中间件
// 使用 OAuth.SetHeader 作为中间件自动签名
func (c *Client) Use(m middleware) {
	c.middleware = append(c.middleware, m)
}

func (c *Client) OAuth(key, secret string) *OAuth {
	c.oauth = NewOAuth(key, secret)
	return c.oauth
}

func (this *Client) Request(method, url string, data interface{}) (reply []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	var (
		req *http.Request
		res *http.Response
	)
	if data != nil {
		var buf []byte
		if buf, err = this.Binder.Marshal(data); err != nil {
			return
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(buf))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	if contentType := this.Binder.String(); contentType != "" {
		req.Header.Add("Content-Type", FormatContentTypeAndCharset(contentType))
	}
	for _, m := range this.middleware {
		if err = m(req); err != nil {
			return
		}
	}
	if this.oauth != nil {
		if err = this.oauth.Request(req); err != nil {
			return
		}
	}

	res, err = http.DefaultClient.Do(req)
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

func (this *Client) Get(url string, reply interface{}) (err error) {
	var body []byte
	body, err = this.Request(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	switch v := reply.(type) {
	case *[]byte:
		*v = body
	case *string:
		*v = string(body)
	default:
		err = this.Binder.Unmarshal(body, reply)
	}

	return
}

func (this *Client) Post(url string, data interface{}, reply interface{}) (err error) {
	var body []byte
	body, err = this.Request(http.MethodPost, url, data)
	if err != nil || reply == nil {
		return
	}
	switch v := reply.(type) {
	case *[]byte:
		*v = body
	case *string:
		*v = string(body)
	default:
		err = this.Binder.Unmarshal(body, reply)
	}
	return
}
