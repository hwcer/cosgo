package express

import (
	stdContext "context"
	"cosgo/app"
	"cosgo/logger"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

type (
	// Engine is the top-level framework instance.
	Engine struct {
		common
		//premiddleware []MiddlewareFunc
		middleware []MiddlewareFunc
		maxParam   *int
		router     *Router
		//notFoundHandler HandlerFunc
		pool   sync.Pool
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

	// Common struct for Engine & Group.
	common struct{}
)

const (
	httpMethodAny = "Any"
	// REPORT Method can be used to get information about a resource, see rfc 3253
	httpMethodREPORT = "REPORT"
	// PROPFIND Method can be used on collection and property resources.
	httpMethodPROPFIND = "PROPFIND"
	charsetUTF8        = "charset=UTF-8"
)

var (
	methods = [...]string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		httpMethodREPORT,
		http.MethodPut,
		http.MethodTrace,
		httpMethodPROPFIND,
	}
)

// Error handlers
var (
	MethodNotFoundHandler = func(c *Context) error {
		return ErrNotFound
	}

	MethodNotAllowedHandler = func(c *Context) error {
		return ErrMethodNotAllowed
	}
)

// New creates an instance of Engine.
func New(address string, tlsConfig ...*tls.Config) (e *Engine) {
	e = &Engine{
		Server:   new(http.Server),
		maxParam: new(int),
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

// DefaultHTTPErrorHandler is the default HTTP error handler. It sends a JSON Response
// with status code.
func (e *Engine) DefaultHTTPErrorHandler(c *Context, err error) {
	if c.Response.committed {
		logger.Warn("%v", err)
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

// CONNECT registers a new CONNECT route for a Path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodConnect, path, h, m...)
}

// DELETE registers a new DELETE route for a Path with matching handler in the router
// with optional route-level middleware.
func (e *Engine) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodDelete, path, h, m...)
}

// GET registers a new GET route for a Path with matching handler in the router
// with optional route-level middleware.
func (e *Engine) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodGet, path, h, m...)
}

// HEAD registers a new HEAD route for a Path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodHead, path, h, m...)
}

// OPTIONS registers a new OPTIONS route for a Path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodOptions, path, h, m...)
}

// PATCH registers a new PATCH route for a Path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodPatch, path, h, m...)
}

// POST registers a new POST route for a Path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodPost, path, h, m...)
}

// PUT registers a new PUT route for a Path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodPut, path, h, m...)
}

// TRACE registers a new TRACE route for a Path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(http.MethodTrace, path, h, m...)
}

// Any registers a new route for all HTTP methods and Path with matching handler
// in the router with optional route-level middleware.
func (e *Engine) Any(path string, h HandlerFunc, m ...MiddlewareFunc) *RNode {
	return e.Add(httpMethodAny, path, h, m...)
}

// matchPath registers a new route for multiple HTTP methods and Path with matching
// handler in the router with optional route-level middleware.
func (e *Engine) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*RNode {
	routes := make([]*RNode, len(methods))
	for i, m := range methods {
		routes[i] = e.Add(m, path, handler, middleware...)
	}
	return routes
}

//TODO
func (e *Engine) RESTful(prefix string, service RESTful) []*RNode {
	routes := make([]*RNode, len(RESTfulMethods))
	return routes
}

// Static registers a new route with Path prefix to serve static files from the
// provided root directory.
func (e *Engine) Static(prefix, root string) *RNode {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}
	return e.static(prefix, root, e.GET)
}

func (common) static(prefix, root string, get func(string, HandlerFunc, ...MiddlewareFunc) *RNode) *RNode {
	h := func(c *Context) error {
		p, err := url.PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}

		name := filepath.Join(root, path.Clean("/"+p)) // "/"+ for security
		fi, err := os.Stat(name)
		if err != nil {
			// The access Path does not exist
			return MethodNotFoundHandler(c)
		}

		// If the Request is for a directory and does not end with "/"
		p = c.Request.URL.Path // Path must not be empty.
		if fi.IsDir() && p[len(p)-1] != '/' {
			// Redirect to ends with "/"
			c.Response.Status(http.StatusMovedPermanently)
			return c.Redirect(p + "/")
		}
		return c.File(name)
	}
	if prefix == "/" {
		return get(prefix+"*", h)
	}
	return get(prefix+"/*", h)
}

func (common) file(path, file string, get func(string, HandlerFunc, ...MiddlewareFunc) *RNode, m ...MiddlewareFunc) *RNode {
	return get(path, func(c *Context) error {
		return c.File(file)
	}, m...)
}

// File registers a new route with Path to serve a static file with optional route-level middleware.
func (e *Engine) File(path, file string, m ...MiddlewareFunc) *RNode {
	return e.file(path, file, e.GET, m...)
}

// Add registers a new route for an HTTP method and Path with matching handler
// in the router with optional route-level middleware.
func (e *Engine) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *RNode {
	name := handlerName(handler)
	r := &RNode{
		MPath:      MPath{Path: path},
		Name:       name,
		Method:     method,
		Handler:    handler,
		middleware: middleware,
	}
	e.router.routes = append(e.router.routes, r)
	return r
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

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire Context
	c := e.AcquireContext(w, r)
	defer func() {
		if err := recover(); err != nil {
			e.HTTPErrorHandler(c, NewHTTPError500(err))
		}
	}()
	c.Path = r.URL.EscapedPath()
	c.matchPath = &MPath{Path: c.Path}
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
	logger.Info("⇨ http(s) server started on %s", e.Server.Addr)
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
	//s.Handler = h2c.NewHandler(e, h2s)
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

//func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
//	for i := len(middleware) - 1; i >= 0; i-- {
//		h = middleware[i](h)
//	}
//	return h
//}
