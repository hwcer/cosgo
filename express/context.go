package express

import (
	"bytes"
	"cosgo/app"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Context struct {
	index  uint8
	query  url.Values
	params map[string]string

	Path     string
	Engine   *Engine
	Routes   []string //当前匹配到的全部路由
	Request  *http.Request
	Response *Response
}

const (
	defaultMemory = 32 << 20 // 32 MB
	indexPage     = "index.html"
	defaultIndent = "  "
)

func (c *Context) writeContentType(value string) {
	header := c.Response.Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

// next should be used only inside middleware.
func (c *Context) next() {
	if c.index >= uint8(len(c.Engine.middleware)) {
		return
	}
	h := c.Engine.middleware[c.index]
	c.index++
	h(c, c.next)
}

func (c *Context) IsTLS() bool {
	return c.Request.TLS != nil
}

func (c *Context) IsWebSocket() bool {
	upgrade := c.Request.Header.Get(HeaderUpgrade)
	return strings.ToLower(upgrade) == "websocket"
}

//协议
func (c *Context) Protocol() string {
	// Can't use `r.Request.URL.Protocol`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}
	if scheme := c.Request.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := c.Request.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := c.Request.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := c.Request.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func (c *Context) RemoteAddr() string {
	if c.Engine != nil && c.Engine.IPExtractor != nil {
		return c.Engine.IPExtractor(c.Request)
	}
	// Fall back to legacy behavior
	if ip := c.Request.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := c.Request.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ra
}

func (c *Context) Param(name string) string {
	return c.params[name]
}

//获取查询参数
func (c *Context) Query(name string) string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query.Get(name)
}

//获取查询字符串
func (c *Context) RawQuery() string {
	return c.Request.URL.RawQuery
}

func (c *Context) FormValue(name string) string {
	return c.Request.FormValue(name)
}

func (c *Context) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.Request.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.Request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.Request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.Request.Form, nil
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	f, fh, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return fh, nil
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(defaultMemory)
	return c.Request.MultipartForm, err
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

func (c *Context) Cookies() []*http.Cookie {
	return c.Request.Cookies()
}

//func (c *Context) Get(key string) interface{} {
//
//}
//
//func (c *Context) Set(key string, val interface{}) {
//
//}

func (c *Context) Bind(i interface{}) error {
	return c.Engine.Binder.Bind(i, c)
}

func (c *Context) Validate(i interface{}) error {
	if c.Engine.Validator == nil {
		return ErrValidatorNotRegistered
	}
	return c.Engine.Validator.Validate(i)
}

func (c *Context) Render(code int, name string, data interface{}) (err error) {
	if c.Engine.Renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.Engine.Renderer.Render(buf, name, data, c); err != nil {
		return
	}
	return c.HTMLBlob(code, buf.Bytes())
}

func (c *Context) HTML(code int, html string) (err error) {
	return c.HTMLBlob(code, []byte(html))
}

func (c *Context) HTMLBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

func (c *Context) String(code int, s string) (err error) {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *Context) jsonPBlob(code int, callback string, i interface{}) (err error) {
	enc := json.NewEncoder(c.Response)
	_, pretty := c.QueryParams()["pretty"]
	if app.Debug || pretty {
		enc.SetIndent("", "  ")
	}
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.Response.WriteHeader(code)
	if _, err = c.Response.Write([]byte(callback + "(")); err != nil {
		return
	}
	if err = enc.Encode(i); err != nil {
		return
	}
	if _, err = c.Response.Write([]byte(");")); err != nil {
		return
	}
	return
}

func (c *Context) json(code int, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	c.writeContentType(MIMEApplicationJSONCharsetUTF8)
	c.Response.Status = code
	return enc.Encode(i)
}

func (c *Context) JSON(code int, i interface{}) (err error) {
	indent := ""
	if _, pretty := c.QueryParams()["pretty"]; app.Debug || pretty {
		indent = defaultIndent
	}
	return c.json(code, i, indent)
}

func (c *Context) JSONPretty(code int, i interface{}, indent string) (err error) {
	return c.json(code, i, indent)
}

func (c *Context) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

func (c *Context) JSONP(code int, callback string, i interface{}) (err error) {
	return c.jsonPBlob(code, callback, i)
}

func (c *Context) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.Response.WriteHeader(code)
	if _, err = c.Response.Write([]byte(callback + "(")); err != nil {
		return
	}
	if _, err = c.Response.Write(b); err != nil {
		return
	}
	_, err = c.Response.Write([]byte(");"))
	return
}

func (c *Context) xml(code int, i interface{}, indent string) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.Response.WriteHeader(code)
	enc := xml.NewEncoder(c.Response)
	if indent != "" {
		enc.Indent("", indent)
	}
	if _, err = c.Response.Write([]byte(xml.Header)); err != nil {
		return
	}
	return enc.Encode(i)
}

func (c *Context) XML(code int, i interface{}) (err error) {
	indent := ""
	if _, pretty := c.QueryParams()["pretty"]; app.Debug || pretty {
		indent = defaultIndent
	}
	return c.xml(code, i, indent)
}

func (c *Context) XMLPretty(code int, i interface{}, indent string) (err error) {
	return c.xml(code, i, indent)
}

func (c *Context) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.Response.WriteHeader(code)
	if _, err = c.Response.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.Response.Write(b)
	return
}

func (c *Context) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.Response.WriteHeader(code)
	_, err = c.Response.Write(b)
	return
}

func (c *Context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.Response.WriteHeader(code)
	_, err = io.Copy(c.Response, r)
	return
}

func (c *Context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return NotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return NotFoundHandler(c)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(c.Response, c.Request, fi.Name(), fi.ModTime(), f)
	return
}

func (c *Context) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *Context) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *Context) contentDisposition(file, name, dispositionType string) error {
	c.Response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

func (c *Context) Empty(code int) error {
	c.Response.WriteHeader(code)
	return nil
}

func (c *Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.Response.Header().Set(HeaderLocation, url)
	c.Response.WriteHeader(code)
	return nil
}

func (c *Context) Error(err error) {
	c.Engine.HTTPErrorHandler(c, err)
}

func (c *Context) Reset(r *http.Request, w http.ResponseWriter) {
	c.index = 0
	c.Request = r
	c.Response.reset(w)
	c.query = nil
	c.store = nil
	c.Path = ""
	c.pnames = nil
	// NOTE: Don't reset because it has to have length c.Engine.maxParam at all times
	for i := 0; i < *c.Engine.maxParam; i++ {
		c.pvalues[i] = ""
	}
}
