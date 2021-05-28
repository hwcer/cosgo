package cosweb

import (
	"github.com/hwcer/cosgo/cosweb/render"
	"io"
)

// Render is the interface that wraps the Render function.
type Render interface {
	Render(io.Writer, string, interface{}) error
}

func NewRender(options *render.Options) *render.Render {
	return render.New(options)
}
