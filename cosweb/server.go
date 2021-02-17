package cosweb

import (
	ctx "context"
	"cosgo/apps"
	"cosgo/logger"
	"crypto/tls"
	"io"
	"net/http"
	"sync"
	"time"
)

type (
	// Server is the top-level framework instance.
	Server struct {
		pool             sync.Pool
		Router           *Router
		middleware       []MiddlewareFunc
		Debug            bool //DEBUG模式会打印所有路由匹配状态,向客户端输出详细错误信息
		Binder           Binder
		Server           *http.Server
		Renderer         Renderer
		HTTPErrorHandler HTTPErrorHandler
	}

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(*Context, func())

	// HandlerFunc defines a function to serve HTTP requests.
	HandlerFunc func(*Context) error

	// HTTPErrorHandler is a centralized HTTP error Handler.
	HTTPErrorHandler func(*Context, error)

	// Renderer is the interface that wraps the Render function.
	Renderer interface {
		Render(io.Writer, string, interface{}, *Context) error
	}
)

var (
	AnyHttpMethod = []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}
	// Error handlers
	MethodNotFoundHandler = func(c *Context) error {
		return ErrNotFound
	}
)

// New creates an instance of Server.
func NewServer(address string, tlsConfig ...*tls.Config) (e *Server) {
	e = &Server{
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
	e.Router = NewRouter()
	return
}

// DefaultHTTPErrorHandler is the default HTTP error Handler. It sends a JSON Response
// with status code.
func (s *Server) DefaultHTTPErrorHandler(c *Context, err error) {
	if c.Response.committed {
		logger.Error(err)
		return
	}

	he, ok := err.(*HTTPError)
	if !ok {
		he = NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.Response.Status(he.Code)

	if c.Request.Method == http.MethodHead {
		c.End()
	} else {
		c.String(err.Error())
	}
}

// Use adds middleware to the chain which is run after Router.
func (s *Server) Use(middleware ...MiddlewareFunc) {
	s.middleware = append(s.middleware, middleware...)
}

// GET registers a new GET Register for a path with matching Handler in the Router
// with optional Register-level middleware.
func (s *Server) GET(path string, h HandlerFunc) {
	s.Register(path, h, http.MethodGet)
}

// POST registers a new POST Register for a path with matching Handler in the
// Router with optional Register-level middleware.
func (s *Server) POST(path string, h HandlerFunc) {
	s.Register(path, h, http.MethodPost)
}

// Any registers a new Register for all HTTP methods and path with matching Handler
// in the Router with optional Register-level middleware.
func (s *Server) Any(path string, h HandlerFunc) {
	s.Register(path, h)
}

// AddTarget registers a new Register for an HTTP value and path with matching Handler
// in the Router with optional Register-level middleware.
func (s *Server) Register(path string, handler HandlerFunc, method ...string) {
	if len(method) == 0 {
		method = AnyHttpMethod
	}
	s.Router.Register(path, handler, method...)
}

//
func (s *Server) Group(prefix string, i interface{}, method ...string) *Group {
	group := NewGroup()
	group.Register(i)
	s.Register(group.Route(prefix), group.handler, method...)
	return group
}

//代理服务器
func (s *Server) Proxy(prefix, address string, method ...string) *Proxy {
	proxy := NewProxy(address)
	s.Register(proxy.Route(prefix), proxy.handle, method...)
	return proxy
}

func (s *Server) RESTful(prefix string, handle iRESTful, method ...string) *RESTful {
	rest := NewRESTful()
	rest.Register(handle)
	if len(method) == 0 {
		method = RESTfulMethods
	}
	s.Register(rest.Route(prefix), rest.handler, method...)
	return rest
}

// Static registers a new Register with path prefix to serve static files from the
// provided root directory.
// 如果root 不是绝对路径 将以程序的WorkDir为基础
func (s *Server) Static(prefix, root string, method ...string) *Static {
	static := NewStatic(root)
	s.Register(static.Route(prefix), static.handler, method...)
	return static
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the Context by calling `ReleaseContext()`.
func (s *Server) AcquireContext(w http.ResponseWriter, r *http.Request) *Context {
	c := s.pool.Get().(*Context)
	if w != nil || r != nil {
		c.reset(r, w)
	}
	return c
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (s *Server) ReleaseContext(c *Context) {
	c.release()
	s.pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire Context
	c := s.AcquireContext(w, r)
	defer func() {
		if err := recover(); err != nil {
			s.HTTPErrorHandler(c, NewHTTPError500(err, s.Debug))
		}
		// Release Context
		s.ReleaseContext(c)
	}()

	c.Path = r.URL.Path
	if len(s.middleware) > 0 {
		c.middleware = append(c.middleware, s.middleware...)
		c.next()
	}
	var err error
	if !c.Aborted() {
		node := s.Router.Match(c.Request.Method, c.Path)
		if node != nil {
			c.Params = node.Params(c.Path)
			if c.Server.Debug {
				logger.Debug("Router matchT success:%v ==> %v", c.Path, node.String())
			}
			if node.Handler != nil {
				err = node.Handler(c)
			}
			if err != nil {
				c.Server.HTTPErrorHandler(c, err)
			}
		}
	}
	//last return 404
	if !c.Response.committed {
		s.HTTPErrorHandler(c, ErrNotFound)
	}
}

// Start starts an HTTP server.
func (s *Server) Start() (err error) {
	err = apps.Timeout(time.Second, func() error {
		if s.Server.TLSConfig != nil {
			return s.Server.ListenAndServeTLS("", "")
		} else {
			return s.Server.ListenAndServe()
		}
	})
	return
}

//立即关闭
func (s *Server) Close() error {
	return s.Server.Close()
}

//优雅关闭，等所有协程结束
func (s *Server) Shutdown(ctx ctx.Context) error {
	return s.Server.Shutdown(ctx)
}
