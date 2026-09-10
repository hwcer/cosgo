package request

import (
	"net/http"
	"strings"
)

const Charset = "UTF-8"

func Address(req *http.Request) (url string) {
	scheme := Protocol(req)
	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	//签名串必须包含path:否则代理/中间人篡改请求路径后验签仍然通过。
	//EscapedPath取percent-encoded形式(与OAuth1规范一致),客户端与验签端
	//都从同一*http.Request取值,两端天然一致
	url = scheme + "://" + host + req.URL.EscapedPath() + "?" + req.URL.RawQuery
	return url
}

func Protocol(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	if scheme := req.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := req.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := req.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := req.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func FormatContentTypeAndCharset(contentType string, charset ...string) string {
	b := strings.Builder{}
	b.WriteString(contentType)
	b.WriteString("; charset=")
	if len(charset) > 0 {
		b.WriteString(charset[0])
	} else {
		b.WriteString(Charset)
	}

	return b.String()
}
