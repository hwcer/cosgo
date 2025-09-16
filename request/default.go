package request

var defaultRequest = New()

func Get(url string, reply interface{}) (err error) {
	return defaultRequest.Get(url, reply)
}

func Post(url string, data interface{}, reply interface{}) (err error) {
	return defaultRequest.Post(url, data, reply)
}

func Request(method, url string, data interface{}, headers ...map[string]string) (reply []byte, err error) {
	return defaultRequest.Request(method, url, data, headers...)
}
