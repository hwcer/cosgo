package cosweb

import "net/http"

func NewCookie(c *Context) *Cookie {
	return &Cookie{c: c}
}

type Cookie struct {
	c *Context
}

func (this *Cookie) Get(key string) string {
	cookie, err := this.c.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (this *Cookie) Set(key, val string) {
	cookie := &http.Cookie{
		Name:  key,
		Value: val,
	}
	http.SetCookie(this.c, cookie)
}

func (this *Cookie) release() {

}

func (this *Cookie) SetCookie(cookie *http.Cookie) {
	http.SetCookie(this.c, cookie)
}
func (this *Cookie) GetCookie(key string) (*http.Cookie, error) {
	return this.c.Request.Cookie(key)
}
