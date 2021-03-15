package cosweb

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
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

//Context API上下文.
type Context struct {
	Path       string
	Server     *Server
	Session    *Session
	Request    *http.Request
	Response   http.ResponseWriter
	query      url.Values
	params     map[string]string
	aborted    bool
	committed  bool
	middleware []MiddlewareFunc
}

// NewContext returns a Context instance.
func NewContext(s *Server, r *http.Request, w http.ResponseWriter) *Context {
	c := &Context{
		Server:   s,
		Request:  r,
		Response: w,
	}
	c.Session = NewSession(s, c)
	return c
}

func (c *Context) reset(r *http.Request, w http.ResponseWriter) {
	c.Session.reset()
	c.Request = r
	c.Response = w
}

//释放资源,准备进入缓存池
func (c *Context) release() {
	c.query = nil
	c.params = nil
	c.aborted = false
	c.committed = false
	c.middleware = nil
	c.Path = ""
	c.Session.release()
	c.Request = nil
	c.Response = nil
}

// next should be used only inside middleware.
func (c *Context) next() {
	if len(c.middleware) == 0 {
		c.aborted = false
	} else {
		handle := c.middleware[0]
		c.middleware = c.middleware[1:]
		handle(c, c.next)
	}
}

//doMiddleware 执行中间件
func (c *Context) doMiddleware(m []MiddlewareFunc) {
	if len(m) > 0 {
		c.aborted = true
		c.middleware = m
		c.next()
	}
}

//IsWebSocket 判断是否WebSocket
func (c *Context) IsWebSocket() bool {
	upgrade := c.Request.Header.Get(HeaderUpgrade)
	return strings.ToLower(upgrade) == "websocket"
}

//Aborted 是否已经被中断
func (c *Context) Aborted() bool {
	return c.aborted
}

//Protocol 协议
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

//RemoteAddr 客户端地址
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

//Get 获取参数,优先路径中的params
//其他方式直接使用c.Request...
func (c *Context) Get(key string, dts ...int) string {
	if len(dts) == 0 {
		dts = defaultGetRequestDataType
	}
	for _, t := range dts {
		if v, ok := GetDataFromRequest(c, key, t); ok {
			return v
		}
	}
	return ""
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

//Bind 绑定JSON XML
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
	c.WriteHeader(0)
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

func (c *Context) JSONP(callback string, i interface{}) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	data = bytes.Join([][]byte{[]byte(callback), []byte("("), data, []byte(")")}, []byte{})
	return c.Bytes(MIMEApplicationJSCharsetUTF8, data)
}

func (c *Context) Bytes(contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	_, err = c.Write(b)
	return
}
func (c *Context) Error(err error) {
	c.Server.HTTPErrorHandler(c, err)
}

func (c *Context) Stream(contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.WriteHeader(0)
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
	http.ServeContent(c, c.Request, fi.Name(), fi.ModTime(), f)
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
	c.WriteHeader(http.StatusMultipleChoices)
	return nil
}

func (c *Context) writeContentType(value string) {
	header := c.Header()
	header.Set(HeaderContentType, value)
}

func (c *Context) contentDisposition(file, name, dispositionType string) error {
	c.Response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}
