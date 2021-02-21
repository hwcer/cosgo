package cosweb

import (
	"cosgo/logger"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

//反向代理服务器

const iProxyRoutePath = "_ProxyRoutePath"

func NewProxy(address ...string) *Proxy {
	proxy := &Proxy{}
	for _, addr := range address {
		proxy.AddTarget(addr)
	}
	proxy.GetTarget = defaultProxyGetTarget
	return proxy
}

type Proxy struct {
	target    []*url.URL
	GetTarget func(*Context, []*url.URL) url.URL //获取目标服务器地址,适用于负载均衡
}

func (this *Proxy) Route(s *Server, prefix string, method ...string) {
	arr := []string{strings.TrimSuffix(prefix, "/"), "*" + iProxyRoutePath}
	r := strings.Join(arr, "/")
	s.Register(r, this.handle, method...)
}

func (this *Proxy) handle(c *Context) error {
	target := this.GetTarget(c, this.target)
	if &target == nil {
		return errors.New("Proxy AddTarget empty")
	}
	path := c.Param(iProxyRoutePath)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	target.Path = path
	target.RawQuery = c.Request.URL.RawQuery
	target.Fragment = c.Request.URL.Fragment
	if c.Request.URL.User != nil {
		target.User = c.Request.URL.User
	}

	address := target.String()

	req, err := http.NewRequest(c.Request.Method, address, c.Request.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	copyHeader(c.Request.Header, &req.Header)

	// Create a client and query the target
	var transport http.Transport
	resp, err := transport.RoundTrip(req)
	if err != nil {
		logger.Error(err)
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	header := c.Response.Header()
	copyHeader(resp.Header, &header)
	header.Add("Requested-Host", req.Host)

	c.Response.WriteHeader(resp.StatusCode)
	c.Response.Write(body)

	return nil
}

//添加代理服务器地址
func (this *Proxy) AddTarget(addr string) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}
	this.target = append(this.target, u)
	return nil
}

func defaultProxyGetTarget(c *Context, address []*url.URL) url.URL {
	var u *url.URL
	if len(address) == 1 {
		u = address[0]
	} else if len(address) > 1 {
		i := rand.Intn(len(address) - 1)
		u = address[i]
	}
	return *u
}

func copyHeader(source http.Header, dest *http.Header) {
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}
