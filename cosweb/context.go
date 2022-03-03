package cosweb

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	indexPage     = "index.html"
	defaultMemory = 32 << 20 // 32 MB
)

type CCPool struct {
	data    interface{}
	reset   func()
	release func()
}

//Context API上下文.
type Context struct {
	pool     *CCPool //缓存逻辑层对象
	body     map[string]interface{}
	query    url.Values
	params   map[string]string
	engine   *Server
	aborted  int
	Cookie   *Cookie
	Session  *Session
	Request  *http.Request
	Response http.ResponseWriter
}

// NewContext returns a Context instance.
func NewContext(s *Server) *Context {
	c := &Context{
		engine: s,
	}
	c.Cookie = NewCookie(c)
	c.Session = NewSession(c)
	return c
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Request = r
	c.Response = w
	//重新设置session id
}

//释放资源,准备进入缓存池
func (c *Context) release() {
	c.body = nil
	c.query = nil
	c.params = nil
	c.aborted = 0
	c.Request = nil
	c.Response = nil
	c.Cookie.release()
	c.Session.release()
}

func (c *Context) next() error {
	c.aborted -= 1
	return nil
}

func (c *Context) doHandle(nodes []*Node) (err error) {
	if len(nodes) == 0 {
		return
	}
	c.aborted += len(nodes)
	num := c.aborted
	for _, node := range nodes {
		num -= 1
		c.params = node.Params(c.Request.URL.Path)
		err = node.Handler(c, c.next)
		if err != nil || c.aborted != num {
			return
		}
	}
	return
}

//doMiddleware 执行中间件
func (c *Context) doMiddleware(middleware []MiddlewareFunc) (error, bool) {
	if len(middleware) == 0 {
		return nil, true
	}
	c.aborted += len(middleware)
	num := c.aborted
	for _, modFun := range middleware {
		num -= 1
		if err := modFun(c, c.next); err != nil {
			return err, false
		}
		if c.aborted != num {
			return nil, false
		}
	}
	return nil, true
}

func (c *Context) Abort() {
	c.aborted += 1
}

//Pool 获取缓存池中缓存的对象
func (c *Context) Pool() (i interface{}) {
	if c.pool != nil {
		i = c.pool.data
	}
	return
}

//IsWebSocket 判断是否WebSocket
func (c *Context) IsWebSocket() bool {
	upgrade := c.Request.Header.Get(HeaderUpgrade)
	return strings.ToLower(upgrade) == "websocket"
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
func (c *Context) Get(key string, dts ...RequestDataType) string {
	if len(dts) == 0 {
		dts = c.engine.RequestDataType
	}
	for _, t := range dts {
		if v, ok := getDataFromRequest(c, key, t); ok {
			return v
		}
	}
	return ""
}
func (c *Context) GetInt(key string, dts ...RequestDataType) int32 {
	v := c.Get(key, dts...)
	if v == "" {
		return 0
	}
	r, _ := strconv.ParseInt(v, 10, 32)
	return int32(r)
}

//Body 将结果快速绑定到Body对象并返回Body
//只绑定BODY(json,xml)内容其他参数通过Get获取
func (c *Context) Body() (body map[string]interface{}, err error) {
	if c.body != nil {
		return c.body, nil
	}
	body = make(map[string]interface{})
	if err = c.Bind(&body); err == nil {
		c.body = body
	}
	return
}

//Bind 绑定JSON XML
func (c *Context) Bind(i interface{}) error {
	return c.engine.Binder.Bind(c, i)
}
