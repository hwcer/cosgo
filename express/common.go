package express

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

type common struct {
}

func (common) static(prefix, root string, get func(string, HandlerFunc, ...MiddlewareFunc) *Route) *Route {
	h := func(c *Context) error {
		p, err := url.PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}

		name := filepath.Join(root, path.Clean("/"+p)) // "/"+ for security
		fi, err := os.Stat(name)
		if err != nil {
			// The access path does not exist
			return MethodNotFoundHandler(c)
		}

		// If the Request is for a directory and does not end with "/"
		p = c.Request.URL.Path // path must not be empty.
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

func (common) file(path, file string, get func(string, HandlerFunc, ...MiddlewareFunc) *Route, m ...MiddlewareFunc) *Route {
	return get(path, func(c *Context) error {
		return c.File(file)
	}, m...)
}
