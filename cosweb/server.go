package cosweb

import (
	ctx "context"
	"crypto/tls"
	"github.com/hwcer/cosgo/cosweb/session"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/utils"
	"golang.org/x/net/context"
	"net/http"
	"sync"
	"time"
)

type (
	// Server is the top-level framework instance.
	Server struct {
		pool             sync.Pool
		Router           *Router
		Binder           Binder
		Render           Render
		Server           *http.Server
		middleware       []MiddlewareFunc
		HTTPErrorHandler HTTPErrorHandler
	}
	// HandlerFunc defines a function to serve HTTP requests.
	HandlerFunc func(*Context, func()) error
	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(*Context, func())
	// HTTPErrorHandler is a centralized HTTP error Handler.
	HTTPErrorHandler func(*Context, error)
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
	he, ok := err.(*HTTPError)
	if !ok {
		he = NewHTTPError(http.StatusInternalServerError, err)
	}

	c.WriteHeader(he.Code)
	if c.Request.Method != http.MethodHead {
		c.String(he.String())
	}
	if Options.Debug {
		logger.Error(he)
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

//Group 注册路由组
func (s *Server) Group(prefix string, i interface{}, method ...string) *Group {
	group := NewGroup()
	group.Register(i)
	group.Route(s, prefix, method...)
	return group
}

//代理服务器
func (s *Server) Proxy(prefix, address string, method ...string) *Proxy {
	proxy := NewProxy(address)
	proxy.Route(s, prefix, method...)
	return proxy
}

func (s *Server) RESTful(prefix string, handle iRESTful, method ...string) *RESTful {
	rest := NewRESTful()
	rest.Register(handle)
	rest.Route(s, prefix, method...)
	return rest
}

// Static registers a new Register with path prefix to serve static files from the
// provided root directory.
// 如果root 不是绝对路径 将以程序的WorkDir为基础
func (s *Server) Static(prefix, root string, method ...string) *Static {
	static := NewStatic(root)
	static.Route(s, prefix, method...)
	return static
}

// AcquireContext returns an empty `Ctx` instance from the pool.
// You must return the Context by calling `ReleaseContext()`.
func (s *Server) AcquireContext(w http.ResponseWriter, r *http.Request) *Context {
	c := s.pool.Get().(*Context)
	if w != nil || r != nil {
		c.reset(r, w)
	}
	return c
}

// ReleaseContext returns the `Ctx` instance back to the pool.
// You must call it after `AcquireContext()`.
func (s *Server) ReleaseContext(c *Context) {
	c.release()
	s.pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := s.AcquireContext(w, r)
	defer func() {
		if err := recover(); err != nil {
			s.HTTPErrorHandler(c, NewHTTPError500(err))
		}
		s.ReleaseContext(c)
	}()
	var err error
	c.doMiddleware(s.middleware)
	if c.aborted == 0 {
		nodes := s.Router.Match(c.Request.Method, c.Request.URL.Path)
		err = c.doHandle(nodes)
	}
	if err != nil {
		s.HTTPErrorHandler(c, err)
	}
}

// Start starts an HTTP server.
func (s *Server) Start() (err error) {
	err = utils.Timeout(time.Second*1, func() error {
		if s.Server.TLSConfig != nil {
			return s.Server.ListenAndServeTLS("", "")
		} else {
			return s.Server.ListenAndServe()
		}
	})
	if err != nil && err != utils.ErrorTimeout {
		return err
	}
	return session.Start(context.Background())
}

//立即关闭
func (s *Server) Close() error {
	err := s.Server.Close()
	if err == nil {
		err = session.Close()
	}
	return err
}

//优雅关闭，等所有协程结束
func (s *Server) Shutdown(ctx ctx.Context) error {
	err := s.Server.Shutdown(ctx)
	if err == nil {
		err = session.Close()
	}
	return err
}
