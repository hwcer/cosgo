package cosweb

import (
	"bytes"
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

const (
	indexPage     = "index.html"
	defaultMemory = 32 << 20 // 32 MB
)

type Context struct {
	query      url.Values
	Path       string
	Server     *Server
	Params     map[string]string
	Request    *http.Request
	Response   *Response
	aborted    bool
	middleware []MiddlewareFunc
}

// context returns a Context instance.
func NewContext(e *Server, r *http.Request, w http.ResponseWriter) *Context {
	return &Context{
		Server:   e,
		Request:  r,
		Response: NewResponse(w, e),
	}
}

func (c *Context) reset(r *http.Request, w http.ResponseWriter) {
	c.Request = r
	c.Response.reset(w)
}

//释放资源,准备进入缓存池
func (c *Context) release() {
	c.Path = ""
	c.query = nil
	c.Params = nil
	c.Request = nil
	c.aborted = false
	c.middleware = nil
	c.Response.release()
}

func (c *Context) writeContentType(value string) {
	header := c.Response.Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

// Next should be used only inside middleware.
func (c *Context) next() {
	if len(c.middleware) == 0 {
		c.aborted = false
	} else {
		c.aborted = true
		handle := c.middleware[0]
		c.middleware = c.middleware[1:]
		handle(c, c.next)
	}
}

func (c *Context) IsWebSocket() bool {
	upgrade := c.Request.Header.Get(HeaderUpgrade)
	return strings.ToLower(upgrade) == "websocket"
}

//是否已经被中断
func (c *Context) Aborted() bool {
	return c.aborted
}

//设置状态码
func (c *Context) Status(code int) *Context {
	c.Response.Status(code)
	return c
}

//协议
func (c *Context) Protocol() string {
	// Can't use `r.Request.URL.Protocol`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.Request.TLS != nil {
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

//获取BODY中的参数
func (c *Context) Get(name string) string {
	return c.Params[name]
}

func (c *Context) Param(name string) string {
	return c.Params[name]
}

//获取查询参数
func (c *Context) Query(name string) string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query.Get(name)
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

func (c *Context) GetCookie(name string) (*http.Cookie, error) {
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
	return c.Server.Binder.Bind(c, i)
}

func (c *Context) Render(name string, data interface{}) (err error) {
	if c.Server.Renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.Server.Renderer.Render(buf, name, data, c); err != nil {
		return
	}
	return c.Bytes(MIMETextHTMLCharsetUTF8, buf.Bytes())
}

//结束响应，返回空内容
func (c *Context) End() error {
	c.Response.WriteHeader(0)
	return nil
}

func (c *Context) XML(i interface{}, indent string) (err error) {
	data, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	c.Bytes(MIMEApplicationXMLCharsetUTF8, data)
	return
}

func (c *Context) HTML(html string) (err error) {
	return c.Bytes(MIMETextHTMLCharsetUTF8, []byte(html))
}

func (c *Context) String(s string) (err error) {
	return c.Bytes(MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *Context) JSON(i interface{}) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.Bytes(MIMEApplicationJSONCharsetUTF8, data)
}

func (c *Context) JSONP(callback string, i interface{}) (err error) {
	enc := json.NewEncoder(c.Response)
	c.writeContentType(MIMEApplicationJSCharsetUTF8)
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

func (c *Context) Bytes(contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	_, err = c.Response.Write(b)
	return
}

func (c *Context) Stream(contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	_, err = io.Copy(c.Response, r)
	return
}

func (c *Context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return MethodNotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return MethodNotFoundHandler(c)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(c.Response, c.Request, fi.Name(), fi.ModTime(), f)
	return
}

func (c *Context) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *Context) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *Context) Redirect(url string) error {
	c.Response.Header().Set(HeaderLocation, url)
	if c.Response.httpStatusCode == 0 {
		c.Response.Status(http.StatusMultipleChoices)
	}
	return nil
}

func (c *Context) contentDisposition(file, name, dispositionType string) error {
	c.Response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

func (c *Context) Error(err error) {
	c.Server.HTTPErrorHandler(c, err)
}
