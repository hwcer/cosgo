package express

import (
	stdContext "context"
	"cosgo/app"
	"cosgo/logger"
	"crypto/tls"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

type (
	// Engine is the top-level framework instance.
	Engine struct {
		common
		//notFoundHandler HandlerFunc
		pool sync.Pool
		//调试模式会打印所有路由匹配状态
		debug  bool
		router *Router
		//premiddleware []MiddlewareFunc
		middleware []MiddlewareFunc

		Server *http.Server
		//TLSServer        *http.Server
		//Listener net.Listener
		//TLSListener      net.Listener
		//AutoTLSManager   autocert.Manager
		//DisableHTTP2     bool

		Binder           Binder
		Validator        Validator
		Renderer         Renderer
		IPExtractor      IPExtractor
		HTTPErrorHandler HTTPErrorHandler
	}

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(*Context)

	// HandlerFunc defines a function to serve HTTP requests.
	HandlerFunc func(*Context) error

	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(*Context, error)

	// Validator is the interface that wraps the Validate function.
	Validator interface {
		Validate(i interface{}) error
	}

	// Renderer is the interface that wraps the Render function.
	Renderer interface {
		Render(io.Writer, string, interface{}, *Context) error
	}

	// Map defines a generic map of type `map[string]interface{}`.
	Map map[string]interface{}
)

const (
	httpMethodAny = "Any"
	// REPORT method can be used to get information about a resource, see rfc 3253
	httpMethodREPORT = "REPORT"
	// PROPFIND method can be used on collection and property resources.
	httpMethodPROPFIND = "PROPFIND"
)

// Error handlers
var (
	MethodNotFoundHandler = func(c *Context) error {
		return ErrNotFound
	}
)

// New creates an instance of Engine.
func New(address string, tlsConfig ...*tls.Config) (e *Engine) {
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

func (e *Engine) Proxy(path string, address ...string) *Route {
	proxy := &Proxy{}
	for _, addr := range address {
		proxy.Add(addr)
	}
	return e.Add(httpMethodAny, path, proxy.handle)
}

// DefaultHTTPErrorHandler is the default HTTP error handler. It sends a JSON Response
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
	if app.Debug {
		message = err.Error()
	} else {
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

// Use adds middleware to the chain which is run after router.
func (e *Engine) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodConnect, path, h, m...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Engine) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodDelete, path, h, m...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Engine) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodGet, path, h, m...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodHead, path, h, m...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodOptions, path, h, m...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodPatch, path, h, m...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodPost, path, h, m...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodPut, path, h, m...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodTrace, path, h, m...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (e *Engine) Any(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(httpMethodAny, path, h, m...)
}

// matchPath registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware.
func (e *Engine) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	return e.router.Match(methods, path, handler, middleware...)
}

//TODO
func (e *Engine) RESTful(prefix string, service RESTful) []*Route {
	routes := make([]*Route, len(RESTfulMethods))
	return routes
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (e *Engine) Static(prefix, root string) *Route {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}
	return e.static(prefix, root, e.GET)
}

// File registers a new route with path to serve a static file with optional route-level middleware.
func (e *Engine) File(path, file string, m ...MiddlewareFunc) *Route {
	return e.file(path, file, e.GET, m...)
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (e *Engine) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	return e.router.Add(method, path, handler, middleware...)
}

// Group creates a new router group with prefix and optional group-level middleware.
func (e *Engine) Group(prefix string, m ...MiddlewareFunc) (g *Group) {
	g = &Group{prefix: prefix, Engine: e}
	g.Use(m...)
	return
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

// ServeHTTP implements `http.handler` interface, which serves HTTP requests.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire Context
	c := e.AcquireContext(w, r)
	defer func() {
		if err := recover(); err != nil {
			e.HTTPErrorHandler(c, NewHTTPError500(err))
		}
	}()
	c.Path = r.URL.Path
	//do middleware
	c.Next()
	if !c.Response.committed {
		e.HTTPErrorHandler(c, ErrNotFound)
	}
	// Release Context
	e.ReleaseContext(c)
}

// Start starts an HTTP server.
func (e *Engine) Start() (err error) {
	err = app.TimeOut(time.Second, func() error {
		if e.Server.TLSConfig != nil {
			return e.Server.ListenAndServeTLS("", "")
		} else {
			return e.Server.ListenAndServe()
		}
	})
	return
}

// StartH2CServer starts a custom http/2 server with h2c (HTTP/2 Cleartext).
func (e *Engine) StartH2(h2s *http2.Server) (err error) {
	// Setup
	//s := e.Server
	//s.handler = h2c.NewHandler(e, h2s)
	//if e.Listener == nil {
	//	e.Listener, err = newListener(s.Addr)
	//	if err != nil {
	//		return err
	//	}
	//}
	//logger.Info("⇨ http2 server started on %s\n", e.address)
	//err = app.TimeOut(time.Second, func() error {
	//	return s.Serve(e.Listener)
	//})
	return
}

// Close immediately stops the server.
// It internally calls `http.Server#Close()`.
func (e *Engine) Close() error {
	return e.Server.Close()
}

// Shutdown stops the server gracefully.
// It internally calls `http.Server#Shutdown()`.
func (e *Engine) Shutdown(ctx stdContext.Context) error {
	return e.Server.Shutdown(ctx)
}
