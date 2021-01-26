package cosweb

import (
	stdContext "context"
	"cosgo/apps"
	"cosgo/logger"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type (
	// Engine is the top-level framework instance.
	Engine struct {
		//notFoundHandler HandlerFunc
		pool sync.Pool

		router *Router
		//premiddleware []MiddlewareFunc
		middleware []MiddlewareFunc

		Server *http.Server
		//TLSServer        *http.Server
		//Listener net.Listener
		//TLSListener      net.Listener
		//AutoTLSManager   autocert.Manager
		//DisableHTTP2     bool

		Debug            bool //DEBUG模式会打印所有路由匹配状态,向客户端输出详细错误信息
		Binder           Binder
		Renderer         Renderer
		HTTPErrorHandler HTTPErrorHandler
	}

	// MiddlewareFunc defines a function to process Middleware.
	MiddlewareFunc func(*Context, func())

	// HandlerFunc defines a function to serve HTTP requests.
	HandlerFunc func(*Context) error

	// HTTPErrorHandler is a centralized HTTP error Handler.
	HTTPErrorHandler func(*Context, error)

	// Renderer is the interface that wraps the Render function.
	Renderer interface {
		Render(io.Writer, string, interface{}, *Context) error
	}

	// Map defines a generic map of type `map[string]interface{}`.
	Map map[string]interface{}
)

const (
	HttpMethodAny = "Any"
	// REPORT value can be used to get information about a resource, see rfc 3253
	HttpMethodREPORT = "REPORT"
	// PROPFIND value can be used on collection and property resources.
	HttpMethodPROPFIND = "PROPFIND"
)

// Error handlers
var (
	MethodNotFoundHandler = func(c *Context) error {
		return ErrNotFound
	}
)

// New creates an instance of Engine.
func NewServer(address string, tlsConfig ...*tls.Config) (e *Engine) {
	e = &Engine{
		Server: new(http.Server),
	}
	if len(tlsConfig) > 0 {
		e.Server.TLSConfig = tlsConfig[0]
	}
	e.Server.Addr = address
	e.Server.Handler = e
	e.HTTPErrorHandler = e.DefaultHTTPErrorHandler
	e.Binder = &DefaultBinder{}
	e.pool.New = func() interface{} {
		return NewContext(e, nil, nil)
	}
	e.router = NewRouter(e)
	return
}

// Router returns the default router.
func (e *Engine) Router() *Router {
	return e.router
}

// DefaultHTTPErrorHandler is the default HTTP error Handler. It sends a JSON Response
// with status code.
func (e *Engine) DefaultHTTPErrorHandler(c *Context, err error) {
	if c.Response.committed {
		logger.Error(err)
		return
	}

	he, ok := err.(*HTTPError)
	if !ok {
		he = &HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	c.Response.Status(he.Code)
	message := ""
	if e.Debug {
		message = err.Error()
	} else {
		logger.Error(err)
		message = http.StatusText(he.Code)
	}
	// Send Response
	if c.Request.Method == http.MethodHead {
		err = c.End()
	} else {
		err = c.String(message)
	}
	if err != nil {
		logger.Error(err)
	}

}

// Use adds Middleware to the chain which is run after router.
func (e *Engine) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// CONNECT registers a new CONNECT Register for a path with matching Handler in the
// router with optional Register-level Middleware.
func (e *Engine) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodConnect}, path, h, m...)
}

// DELETE registers a new DELETE Register for a path with matching Handler in the router
// with optional Register-level Middleware.
func (e *Engine) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodDelete}, path, h, m...)
}

// GET registers a new GET Register for a path with matching Handler in the router
// with optional Register-level Middleware.
func (e *Engine) GET(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodGet}, path, h, m...)
}

// HEAD registers a new HEAD Register for a path with matching Handler in the
// router with optional Register-level Middleware.
func (e *Engine) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodHead}, path, h, m...)
}

// OPTIONS registers a new OPTIONS Register for a path with matching Handler in the
// router with optional Register-level Middleware.
func (e *Engine) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodOptions}, path, h, m...)
}

// PATCH registers a new PATCH Register for a path with matching Handler in the
// router with optional Register-level Middleware.
func (e *Engine) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodPatch}, path, h, m...)
}

// POST registers a new POST Register for a path with matching Handler in the
// router with optional Register-level Middleware.
func (e *Engine) POST(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodPost}, path, h, m...)
}

// PUT registers a new PUT Register for a path with matching Handler in the
// router with optional Register-level Middleware.
func (e *Engine) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodPut}, path, h, m...)
}

// TRACE registers a new TRACE Register for a path with matching Handler in the
// router with optional Register-level Middleware.
func (e *Engine) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{http.MethodTrace}, path, h, m...)
}

// Any registers a new Register for all HTTP methods and path with matching Handler
// in the router with optional Register-level Middleware.
func (e *Engine) Any(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.Register([]string{HttpMethodAny}, path, h, m...)
}

// AddTarget registers a new Register for an HTTP value and path with matching Handler
// in the router with optional Register-level Middleware.
func (e *Engine) Register(method []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	e.router.Register(method, path, handler, middleware...)
}

//
func (e *Engine) Group(prefix string, i interface{}, middleware ...MiddlewareFunc) *Group {
	arr := []string{strings.TrimSuffix(prefix, "/"), ":" + iGroupRoutePath, ":" + iGroupRouteName}
	nsp := NewGroup()
	nsp.Register(i)
	e.router.Register([]string{HttpMethodAny}, strings.Join(arr, "/"), nsp.handler, middleware...)
	return nsp
}

//代理服务器
func (e *Engine) Proxy(prefix, address string, middleware ...MiddlewareFunc) *Proxy {
	arr := []string{strings.TrimSuffix(prefix, "/"), iProxyRoutePath}
	proxy := NewProxy(address)
	e.router.Register([]string{HttpMethodAny}, strings.Join(arr, "/"), proxy.handle, middleware...)
	return proxy
}

func (e *Engine) RESTful(prefix string, handle iRESTful, middleware ...MiddlewareFunc) *RESTful {
	arr := []string{strings.TrimSuffix(prefix, "/"), ":" + iRESTfulRoutePath}
	rest := NewRESTful()
	rest.Register(handle)
	method := append([]string{}, RESTfulMethods...)
	e.router.Register(method, strings.Join(arr, "/"), rest.handler, middleware...)
	return rest
}

// Static registers a new Register with path prefix to serve static files from the
// provided root directory.
// 如果root 不是绝对路径 将以程序的WorkDir为基础
func (e *Engine) Static(prefix, root string, middleware ...MiddlewareFunc) *Static {
	arr := []string{strings.TrimSuffix(prefix, "/"), iStaticRoutePath}
	static := NewStatic(root)
	method := append([]string{}, HttpMethodAny)
	e.router.Register(method, strings.Join(arr, "/"), static.handler, middleware...)
	return static
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the Context by calling `ReleaseContext()`.
func (e *Engine) AcquireContext(w http.ResponseWriter, r *http.Request) *Context {
	c := e.pool.Get().(*Context)
	if w != nil || r != nil {
		c.reset(r, w)
	}
	return c
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (e *Engine) ReleaseContext(c *Context) {
	e.pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire Context
	c := e.AcquireContext(w, r)
	defer func() {
		if err := recover(); err != nil {
			e.HTTPErrorHandler(c, NewHTTPError500(err))
		}
		// Release Context
		e.ReleaseContext(c)
	}()

	c.Path = r.URL.Path
	c.middleware.reset(e.middleware...)
	//do Middleware
	c.next()
	var err error
	if !c.Aborted() {
		node := e.router.Match(c.Path, c.Request.Method)
		if node != nil && node.Handler != nil {
			if c.Engine.Debug {
				logger.Debug("router match success:%v ==> %v", c.Path, node.String())
			}
			if len(node.Middleware) > 0 {
				c.middleware.reset(node.Middleware...)
				c.next()
			}
			if !c.Aborted() {
				err = node.Handler(c)
			}
			if err != nil {
				c.Engine.HTTPErrorHandler(c, err)
			}
		}
	}
	if !c.Response.committed {
		e.HTTPErrorHandler(c, ErrNotFound)
	}
}

// Start starts an HTTP server.
func (e *Engine) Start() (err error) {
	err = apps.Timeout(time.Second, func() error {
		if e.Server.TLSConfig != nil {
			return e.Server.ListenAndServeTLS("", "")
		} else {
			return e.Server.ListenAndServe()
		}
	})
	return
}

//立即关闭
func (e *Engine) Close() error {
	return e.Server.Close()
}

//优雅关闭，等所有协程结束
func (e *Engine) Shutdown(ctx stdContext.Context) error {
	return e.Server.Shutdown(ctx)
}
