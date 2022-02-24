package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var c = &http.Client{Timeout: time.Second * 10}

func New() *Client {
	r := &Client{}
	r.Packer = &PackerJson{}
	return r
}

type Client struct {
	Packer Packer
	Header func(method, url string, data io.Reader) map[string]string
}

func (this *Client) reader(i interface{}) (rd io.Reader, err error) {
	if i == nil {
		return nil, err
	}
	if v, ok := i.(io.Reader); ok {
		return v, nil
	}
	switch i.(type) {
	case string:
		rd = bytes.NewReader([]byte(i.(string)))
	case []byte:
		rd = bytes.NewReader(i.([]byte))
	default:
		bf := new(bytes.Buffer)
		err = this.Packer.Encode(bf, i)
		rd = bf
	}
	return
}

func (this *Client) Request(method, url string, data interface{}, reply func(io.Reader) error) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	var (
		req *http.Request
		res *http.Response
	)
	var buf io.Reader
	if buf, err = this.reader(data); err != nil {
		return
	}

	req, err = http.NewRequest(method, url, buf)
	if err != nil {
		return
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	if contentType := this.Packer.ContentType(); contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	if this.Header != nil {
		for k, v := range this.Header(method, url, buf) {
			req.Header.Add(k, v)
		}
	}
	res, err = c.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	if reply != nil {
		err = reply(res.Body)
	}
	return
}

func (this *Client) Get(url string, reply interface{}) (err error) {
	return this.Request(http.MethodGet, url, nil, func(reader io.Reader) error {
		return this.Packer.Decode(reader, reply)
	})
}

func (this *Client) Post(url string, data interface{}, reply interface{}) (err error) {
	return this.Request(http.MethodPost, url, data, func(reader io.Reader) error {
		return this.Packer.Decode(reader, reply)
	})
}
