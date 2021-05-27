package cosweb

import (
	"html/template"
	"io"
	"path/filepath"
)

// Renderer is the interface that wraps the Render function.
type Renderer interface {
	Render(io.Writer, string, interface{}) error
}

type DefaultRenderer struct {
	root []string
}

func (r *DefaultRenderer) AddPath(path string) {
	r.root = append(r.root, path)
}

func (r *DefaultRenderer) Render(buf io.Writer, name string, data interface{}) error {
	var arr []string
	for _, v := range r.root {
		arr = append(arr, filepath.Join(v, name))
	}
	t, err := template.ParseFiles(arr...)
	if err != nil {
		return err
	}
	return t.Execute(buf, data)
}
