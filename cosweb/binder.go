package cosweb

import (
	"encoding/json"
	"encoding/xml"
	"strings"
)

type (
	// Binder is the interface that wraps the Bind value.
	Binder interface {
		Bind(c *Context, i interface{}) error
	}
	// DefaultBinder is the default implementation of the Binder interface.
	DefaultBinder struct{}
)

// Bind implements the `Packer#Bind` function.
func (b *DefaultBinder) Bind(c *Context, i interface{}) (err error) {
	req := c.Request
	if req.ContentLength == 0 {
		return
	}
	ctype := strings.ToLower(req.Header.Get(HeaderContentType))
	switch {
	case strings.HasPrefix(ctype, ContentTypeApplicationJSON):
		return json.NewDecoder(req.Body).Decode(i)
	case strings.HasPrefix(ctype, ContentTypeApplicationXML), strings.HasPrefix(ctype, ContentTypeTextXML):
		return xml.NewDecoder(req.Body).Decode(i)
	default:
		return ErrUnsupportedMediaType
	}
}
