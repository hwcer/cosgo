package middleware

import (
	"github.com/hwcer/cosgo/cosweb"
	"net/http"
	"strconv"
	"strings"
)

/*跨域
access := middleware.NewAccessControlAllow("www.test.com","www.test1.com")
cosweb.Use(access.handle)
*/

type AccessControlAllow struct {
	expire      string
	origin      []string
	methods     []string
	headers     []string
	Credentials bool
}

func NewAccessControlAllow(origin ...string) *AccessControlAllow {
	return &AccessControlAllow{
		origin: origin,
	}
}

func NewAccessControlAllowHandle(origin ...string) cosweb.MiddlewareFunc {
	aca := NewAccessControlAllow(origin...)
	return aca.Handle
}

func (this *AccessControlAllow) Expire(second int) {
	this.expire = strconv.Itoa(second)
}
func (this *AccessControlAllow) Origin(origin ...string) {
	this.origin = append(this.origin, origin...)
}
func (this *AccessControlAllow) Methods(methods ...string) {
	this.methods = append(this.methods, methods...)
}
func (this *AccessControlAllow) Headers(headers ...string) {
	this.headers = append(this.headers, headers...)
}

func (this *AccessControlAllow) Handle(c *cosweb.Context, next cosweb.Next) error {
	header := c.Header()

	if len(this.origin) > 0 {
		header.Add("Access-Control-Allow-Origin", strings.Join(this.origin, ","))
	}
	if len(this.methods) > 0 {
		header.Add("Access-Control-Allow-Methods", strings.Join(this.methods, ","))
	}
	if len(this.headers) > 0 {
		header.Add("Access-Control-Allow-Headers", strings.Join(this.headers, ","))
	}
	if this.Credentials {
		header.Set("Access-Control-Allow-Credentials", "true")
	}
	if this.expire != "" {
		header.Set("Access-Control-Max-Age", this.expire)
	}
	if c.Request.Method == http.MethodOptions {
		c.Write([]byte("options OK"))
	} else {
		next()
	}
	return nil
}
