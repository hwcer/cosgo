package express

import (
	"strings"
)

type RNode struct {
	MPath
	Path       string
	Name       string
	Method     string
	Handler    HandlerFunc
	middleware []MiddlewareFunc
}

type MPath struct {
	Path      string
	matchPath []string
}

// Router is the registry of all registered Routes for an `Engine` instance for
// Request matching and URL Path parameter parsing.
type Router struct {
	engine *Engine
	routes []*RNode
}

func (m *MPath) GetMatchPath() []string {
	if len(m.matchPath) == 0 {
		m.matchPath = strings.Split(m.Path, "/")
	}
	return m.matchPath
}

func (r *RNode) Match(method string, tar *MPath) (param map[string]string, ok bool) {
	param = make(map[string]string)
	if method != r.Method {
		return
	}
	if r.Path == tar.Path {
		return param, true
	}

	routeMatch := r.GetMatchPath()
	targetMatch := tar.GetMatchPath()
	rl := len(routeMatch)
	tl := len(targetMatch)
	if rl > tl || (rl < tl && routeMatch[rl-1] != "*") {
		return
	}
	for i := 0; i < rl; i++ {
		if routeMatch[i] == "*" {
			continue
		} else if strings.HasPrefix(routeMatch[i], ":") {
			k := strings.TrimLeft(routeMatch[i], ":")
			param[k] = targetMatch[i]
		} else if routeMatch[i] != targetMatch[i] {
			return nil, false
		}
	}
	return param, true
}

// NewRouter returns a new Router instance.
func NewRouter(e *Engine) *Router {
	return &Router{
		engine: e,
		routes: make([]*RNode, 0),
	}
}

// Add registers a new route for method and Path with matching handler.
func (r *Router) Add(method, path string, h HandlerFunc) {

}

// Find lookup a handler registered for method and Path. It also parses URL for Path
// parameters and load them into Context.
//
// For performance:
//
// - Get Context from `Engine#AcquireContext()`
// - reset it `Context#reset()`
// - Return it `Engine#ReleaseContext()`.

func (r *Router) Find(c *Context, method, path string) {

}
