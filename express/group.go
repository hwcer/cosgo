package express

import (
	"net/http"
)

type (
	// Group is a set of sub-Routes for a specified route. It can be used for inner
	// Routes that share a common middleware or functionality that should be separate
	// from the parent Engine instance while still inheriting from it.
	Group struct {
		common
		host       string
		prefix     string
		middleware []MiddlewareFunc
		Engine     *Engine
	}
)

// Use implements `Engine#Use()` for sub-Routes within the Group.
func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
	if len(g.middleware) == 0 {
		return
	}
	// Allow all requests to reach the group as they might get dropped if router
	// doesn't find a match, making none of the group middleware process.
}

// CONNECT implements `Engine#CONNECT()` for sub-Routes within the Group.
func (g *Group) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodConnect, path, h, m...)
}

// DELETE implements `Engine#DELETE()` for sub-Routes within the Group.
func (g *Group) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodDelete, path, h, m...)
}

// GET implements `Engine#GET()` for sub-Routes within the Group.
func (g *Group) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodGet, path, h, m...)
}

// HEAD implements `Engine#HEAD()` for sub-Routes within the Group.
func (g *Group) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodHead, path, h, m...)
}

// OPTIONS implements `Engine#OPTIONS()` for sub-Routes within the Group.
func (g *Group) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodOptions, path, h, m...)
}

// PATCH implements `Engine#PATCH()` for sub-Routes within the Group.
func (g *Group) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodPatch, path, h, m...)
}

// POST implements `Engine#POST()` for sub-Routes within the Group.
func (g *Group) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodPost, path, h, m...)
}

// PUT implements `Engine#PUT()` for sub-Routes within the Group.
func (g *Group) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodPut, path, h, m...)
}

// TRACE implements `Engine#TRACE()` for sub-Routes within the Group.
func (g *Group) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(http.MethodTrace, path, h, m...)
}

// Any implements `Engine#Any()` for sub-Routes within the Group.
//func (g *Group) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route {
//	route := make([]*Route, len(methods))
//	for i, m := range methods {
//		route[i] = g.Add(m, path, handler, middleware...)
//	}
//	return route
//}

// matchPath implements `Engine#matchPath()` for sub-Routes within the Group.
func (g *Group) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route {
	routes := make([]*Route, len(methods))
	for i, m := range methods {
		routes[i] = g.Add(m, path, handler, middleware...)
	}
	return routes
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) (sg *Group) {
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	sg = g.Engine.Group(g.prefix+prefix, m...)
	sg.host = g.host
	return
}

// Static implements `Engine#Static()` for sub-Routes within the Group.
func (g *Group) Static(prefix, root string) {
	g.static(prefix, root, g.GET)
}

// File implements `Engine#File()` for sub-Routes within the Group.
func (g *Group) File(path, file string) {
	g.file(path, file, g.GET)
}

// Add implements `Engine#Add()` for sub-Routes within the Group.
func (g *Group) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	// Combine into a new slice to avoid accidentally passing the same slice for
	// multiple Routes, which would lead to later add() calls overwriting the
	// middleware from earlier calls.
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	return g.Engine.Add(method, g.prefix+path, handler, m...)
}
