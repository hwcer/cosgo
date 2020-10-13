package express

import (
	"cosgo/logger"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
)

//反向代理服务器

type Proxy struct {
	targets      []string
	reverseProxy *httputil.ReverseProxy
}

func (this *Proxy) handle(c *Context) error {
	target := this.targets[0]
	return report(c.Response.Writer, c.Request, target)
}

func (this *Proxy) Add(url string) {
	this.targets = append(this.targets, url)
}

//func (this *Proxy) NewMultipleHostsReverseProxy() *httputil.ReverseProxy {
//	//return httputil.NewSingleHostReverseProxy(this.targets[0])
//	if this.reverseProxy == nil {
//		this.reverseProxy = &httputil.ReverseProxy{Director: this.director}
//	}
//	return this.reverseProxy
//}

//func (this *Proxy) director(req *http.Request) {
//	target := this.targets[0]
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

func report(w http.ResponseWriter, r *http.Request, url string) error {

	uri := url + r.RequestURI

	logger.Debug(r.Method + ": " + uri)

	rr, err := http.NewRequest(r.Method, uri, r.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	copyHeader(r.Header, &rr.Header)

	// Create a client and query the target
	var transport http.Transport
	resp, err := transport.RoundTrip(rr)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("Resp-Headers: %v", resp.Header)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	dH := w.Header()
	copyHeader(resp.Header, &dH)
	dH.Add("Requested-Host", rr.Host)

	w.Write(body)

	return nil
}

func copyHeader(source http.Header, dest *http.Header) {
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}
