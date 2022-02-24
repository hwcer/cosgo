package request

import "io"

var defaultRequest = New()

func Get(url string, reply interface{}) (err error) {
	return defaultRequest.Get(url, reply)
}

func Post(url string, data interface{}, reply interface{}) (err error) {
	return defaultRequest.Post(url, data, reply)
}

func Request(method, url string, data interface{}, reply func(io.Reader) error) (err error) {
	return defaultRequest.Request(method, url, data, reply)
}
