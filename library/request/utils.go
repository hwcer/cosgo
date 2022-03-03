package request

import (
	"net/http"
	"strings"
)

func RequestURL(req *http.Request) (url string) {
	scheme := req.URL.Scheme
	if scheme == "" {
		i := strings.Index(req.Proto, "/")
		scheme = strings.ToLower(req.Proto[0:i])
	}
	host := req.URL.Host
	if host == "" {
		host = req.Host
	}

	url = scheme + "://" + host + req.RequestURI
	return
}
