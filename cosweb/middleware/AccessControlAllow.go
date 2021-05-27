package middleware

import (
	"github.com/hwcer/cosgo/cosweb"
	"net/http"
	"strconv"
	"strings"
)

/*跨域
access := middleware.NewAccessControlAllow("www.test.com","www.test1.com")
cosweb.Use(access.Handle)
*/

type AccessControlAllow struct {
	MaxAge      int
	Origin      []string
	Methods     []string
	Headers     []string
	Credentials bool
}

func NewAccessControlAllow(origin ...string) *AccessControlAllow {
	return &AccessControlAllow{
		Origin: origin,
	}
}

func NewAccessControlAllowHandle(origin ...string) cosweb.MiddlewareFunc {
	aca := NewAccessControlAllow(origin...)
	return aca.Handle
}

func (this *AccessControlAllow) Handle(c *cosweb.Context) error {
	header := c.Header()

	if len(this.Origin) > 0 {
		header.Add("Access-Control-Allow-Origin", strings.Join(this.Origin, ","))
	}
	if len(this.Methods) > 0 {
		header.Add("Access-Control-Allow-Methods", strings.Join(this.Methods, ","))
	}
	if len(this.Headers) > 0 {
		header.Add("Access-Control-Allow-Headers", strings.Join(this.Headers, ","))
	}
	if this.Credentials {
		header.Set("Access-Control-Allow-Credentials", "true")
	}
	if this.MaxAge > 0 {
		header.Set("Access-Control-Max-Age", strconv.Itoa(this.MaxAge))
	}
	if c.Request.Method == http.MethodOptions {
		c.HTML("options OK")
	}
	return nil
}
