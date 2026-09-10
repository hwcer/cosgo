package request

var defaultRequest = New()

func Get(url string, reply any) (err error) {
	return defaultRequest.Get(url, reply)
}

func Post(url string, data any, reply any) (err error) {
	return defaultRequest.Post(url, data, reply)
}

func Request(method, url string, data any, headers ...map[string]string) (reply []byte, err error) {
	return defaultRequest.Request(method, url, data, headers...)
}
