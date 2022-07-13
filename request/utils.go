package request

import (
	"net/http"
)

func Address(req *http.Request) (url string) {
	scheme := Protocol(req)
	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	url = scheme + "://" + host + "?" + req.URL.RawQuery
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
