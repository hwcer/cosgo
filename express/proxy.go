package express

import (
	"cosgo/logger"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
)

//反向代理服务器

func NewProxy(address ...string) *Proxy {
	proxy := &Proxy{}
	for _, addr := range address {
		proxy.SetAddress(addr)
	}
	proxy.GetPath = defaultProxyGetPath
	proxy.GetAddress = defaultProxyGetAddress
	return proxy
}

type Proxy struct {
	address    []string
	GetPath    func(*Context) string           //获取客户端请求PATH,需要路径重写时使用
	GetAddress func(*Context, []string) string //获取目标服务器地址,适用于负载均衡
}

func (this *Proxy) handle(c *Context) error {
	url := this.GetAddress(c, this.address)
	if url == "" {
		return errors.New("Proxy Address empty")
	}
	path := this.GetPath(c)

	address := url + path

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

	c.Response.Status(resp.StatusCode)
	c.Response.Write(body)

	return nil
}

//添加代理服务器地址
func (this *Proxy) SetAddress(url string) {
	if strings.HasPrefix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	this.address = append(this.address, url)
}

//func (this *Proxy) NewMultipleHostsReverseProxy() *httputil.ReverseProxy {
//	//return httputil.NewSingleHostReverseProxy(this.address[0])
//	if this.reverseProxy == nil {
//		this.reverseProxy = &httputil.ReverseProxy{Director: this.director}
//	}
//	return this.reverseProxy
//}

//func (this *Proxy) director(req *http.Request) {
//	target := this.address[0]
//	targetQuery := target.RawQuery
//
//	req.Host = target.Host
//	req.URL.Scheme = target.Scheme
//	req.URL.Host = target.Host
//	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
//	if targetQuery == "" || req.URL.RawQuery == "" {
//		req.URL.RawQuery = targetQuery + req.URL.RawQuery
//	} else {
//		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
//	}
//	if _, ok := req.Header["User-Agent"]; !ok {
//		// explicitly disable User-Agent so it's not set to default value
//		req.Header.Set("User-Agent", "")
//	}
//}

func defaultProxyGetPath(c *Context) string {
	return c.Path
}

func defaultProxyGetAddress(c *Context, address []string) string {
	if len(address) == 0 {
		return ""
	} else if len(address) == 1 {
		return address[0]
	}
	i := rand.Intn(len(address) - 1)
	return address[i]
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func copyHeader(source http.Header, dest *http.Header) {
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}
