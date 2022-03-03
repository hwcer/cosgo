package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func New() *Client {
	r := &Client{}
	r.Packer = &PackerJson{}
	return r
}

type middleware func(req *http.Request) error

type Client struct {
	Packer     Packer
	middleware []middleware
}

func (c *Client) Use(m middleware) {
	c.middleware = append(c.middleware, m)
}

func (this *Client) reader(i interface{}) (rd []byte, err error) {
	if i == nil {
		return nil, err
	}
	switch i.(type) {
	case string:
		rd = []byte(i.(string))
	case []byte:
		rd = i.([]byte)
	default:
		rd, err = this.Packer.Encode(i)
	}
	return
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
	var buf []byte
	if buf, err = this.reader(data); err != nil {
		return
	}

	req, err = http.NewRequest(method, url, bytes.NewReader(buf))
	if err != nil {
		return
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	if contentType := this.Packer.ContentType(); contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	for _, m := range this.middleware {
		if err = m(req); err != nil {
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
	err = this.Packer.Decode(body, reply)
	return
}

func (this *Client) Post(url string, data interface{}, reply interface{}) (err error) {
	var body []byte
	body, err = this.Request(http.MethodPost, url, data)
	if err != nil {
		return
	}
	err = this.Packer.Decode(body, reply)
	return
}
