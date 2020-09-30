package express

import (
	"bytes"
	stdContext "context"
	"cosgo/app"
	"cosgo/logger"
	"crypto/tls"
	"fmt"
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
		routers    map[string]*Router
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
	MiddlewareFunc func(*Context, func())

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

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

const (
	charsetUTF8 = "charset=UTF-8"
	// PROPFIND Method can be used on collection and property resources.
	PROPFIND = "PROPFIND"
	// REPORT Method can be used to get information about a resource, see rfc 3253
	REPORT = "REPORT"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Protocol"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderReferrerPolicy                  = "Referrer-Policy"
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
		PROPFIND,
		http.MethodPut,
		http.MethodTrace,
		REPORT,
	}
)

// Error handlers
var (
	NotFoundHandler = func(c *Context) error {
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
		return e.NewContext(nil, nil)
	}
	e.router = NewRouter(e)
	e.routers = map[string]*Router{}
	return
}

// NewContext returns a Context instance.
func (e *Engine) NewContext(r *http.Request, w http.ResponseWriter) *Context {
	return &Context{
		Request:  r,
		Response: NewResponse(w, e),
		store:    make(Map),
		Engine:   e,
		pvalues:  make([]string, *e.maxParam),
	}
}

// Router returns the default router.
func (e *Engine) Router() *Router {
	return e.router
}

// Routers returns the map of host => router.
func (e *Engine) Routers() map[string]*Router {
	return e.routers
}

// DefaultHTTPErrorHandler is the default HTTP error handler. It sends a JSON Response
// with status code.
func (e *Engine) DefaultHTTPErrorHandler(c *Context, err error) {
	he, ok := err.(*HTTPError)
	if !ok {
		he = &HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	code := he.Code
	message := ""
	if app.Debug {
		message = err.Error()
	} else {
		message = http.StatusText(code)
	}
	// Send Response
	if !c.Response.Committed {
		if c.Request.Method == http.MethodHead { // Issue #608
			err = c.Empty(he.Code)
		} else {
			err = c.String(code, message)
		}
		if err != nil {
			logger.Error(err)
		}
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
func (e *Engine) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*RNode {
	routes := make([]*RNode, len(methods))
	for i, m := range methods {
		routes[i] = e.Add(m, path, handler, middleware...)
	}
	return routes
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
			return NotFoundHandler(c)
		}

		// If the Request is for a directory and does not end with "/"
		p = c.Request.URL.Path // Path must not be empty.
		if fi.IsDir() && p[len(p)-1] != '/' {
			// Redirect to ends with "/"
			return c.Redirect(http.StatusMovedPermanently, p+"/")
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

func (e *Engine) add(host, method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *RNode {
	name := handlerName(handler)
	//router := e.findRouter(host)
	//router.Add(method, MPath, func(c *Context) error {
	//	h := applyMiddleware(handler, middleware...)
	//	return h(c)
	//})
	r := &RNode{
		Path:   path,
		Name:   name,
		Method: method,
	}
	e.router.routes[method+path] = r
	return r
}

// Add registers a new route for an HTTP method and Path with matching handler
// in the router with optional route-level middleware.
func (e *Engine) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *RNode {
	return e.add("", method, path, handler, middleware...)
}

// Host creates a new router group for the provided host and optional host-level middleware.
func (e *Engine) Host(name string, m ...MiddlewareFunc) (g *Group) {
	e.routers[name] = NewRouter(e)
	g = &Group{host: name, echo: e}
	g.Use(m...)
	return
}

// Group creates a new router group with prefix and optional group-level middleware.
func (e *Engine) Group(prefix string, m ...MiddlewareFunc) (g *Group) {
	g = &Group{prefix: prefix, echo: e}
	g.Use(m...)
	return
}

// URI generates a URI from handler.
func (e *Engine) URI(handler HandlerFunc, params ...interface{}) string {
	name := handlerName(handler)
	return e.Reverse(name, params...)
}

// URL is an alias for `URI` function.
func (e *Engine) URL(h HandlerFunc, params ...interface{}) string {
	return e.URI(h, params...)
}

// Reverse generates an URL from route name and provided parameters.
func (e *Engine) Reverse(name string, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for _, r := range e.router.routes {
		if r.Name == name {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < ln {
					for ; i < l && r.Path[i] != '/'; i++ {
					}
					uri.WriteString(fmt.Sprintf("%v", params[n]))
					n++
				}
				if i < l {
					uri.WriteByte(r.Path[i])
				}
			}
			break
		}
	}
	return uri.String()
}

// Routes returns the registered Routes.
func (e *Engine) Routes() []*RNode {
	routes := make([]*RNode, 0, len(e.router.routes))
	for _, v := range e.router.routes {
		routes = append(routes, v)
	}
	return routes
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the Context by calling `ReleaseContext()`.
func (e *Engine) AcquireContext() *Context {
	return e.pool.Get().(*Context)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (e *Engine) ReleaseContext(c *Context) {
	e.pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire Context
	c := e.pool.Get().(*Context)
	c.Reset(r, w)

	defer func() {
		if err := recover(); err != nil {
			e.HTTPErrorHandler(c, NewHTTPError500(err))
		}
	}()
	//do middleware
	c.next()
	//Find Router
	h := NotFoundHandler
	//e.findRouter(r.Host).Find(r.Method, r.URL.EscapedPath(), c)
	//
	//h = c.Handler()
	//h = applyMiddleware(h, e.middleware...)

	// Execute chain
	if err := h(c); err != nil {
		e.HTTPErrorHandler(c, err)
	}

	// Release Context
	e.pool.Put(c)
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

func (e *Engine) findRouter(host string) *Router {
	if len(e.routers) > 0 {
		if r, ok := e.routers[host]; ok {
			return r
		}
	}
	return e.router
}

//func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
//	for i := len(middleware) - 1; i >= 0; i-- {
//		h = middleware[i](h)
//	}
//	return h
//}
