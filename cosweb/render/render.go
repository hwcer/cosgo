package render

import (
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	"html/template"
	"io"
	"path/filepath"
)

var funcs template.FuncMap

func init() {
	funcs = make(template.FuncMap)
	funcs["unescaped"] = unescaped
}
func unescaped(x string) interface{} {
	return template.URL(x)
}

// Render provides functions for easily writing HTML templates & JSON out to a HTTP Response.
type Render struct {
	Options   *Options
	Templates *templates
}

// Options holds the configuration Options for a Render
type Options struct {
	Ext string
	// With Debug set to true, templates will be recompiled before every render call.
	Debug bool
	// The glob string to your templates
	Includes  string
	Templates string

	// The Glob string for additional templates
	//PartialsGlob string

	// The function map to pass to each HTML template
	Funcs template.FuncMap

	// Charset for responses
	Charset string

	Delims []string
}

// New creates a new Render with the given Options
func New(opts *Options) *Render {
	if opts == nil {
		opts = &Options{}
	}
	if opts.Ext == "" {
		opts.Ext = ".html"
	}
	if opts.Funcs == nil {
		opts.Funcs = make(template.FuncMap)
	}
	for k, v := range funcs {
		if _, ok := opts.Funcs[k]; !ok {
			opts.Funcs[k] = v
		}
	}

	if opts.Charset == "" {
		opts.Charset = "UTF-8"
	}

	r := &Render{
		Options:   opts,
		Templates: &templates{},
	}

	r.compileTemplatesFromDir()
	return r
}

// HTML executes the template and writes to the responsewriter
func (r *Render) Render(buf io.Writer, name string, data interface{}) error {
	// re-compile on every render call when Debug is true
	name += r.Options.Ext
	if r.Options.Debug {
		r.compileTemplatesFromDir()
	}
	tmpl := r.Templates.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("unrecognised template %s", name)
	}
	// execute template
	err := tmpl.Execute(buf, data)
	if err != nil {
		return err
	}
	return nil
}

func (r *Render) compileTemplatesFromDir() {
	if r.Options.Templates == "" {
		return
	}

	// replace existing templates.
	// NOTE: this is unsafe, but Debug should really not be true in production environments.
	templateSet := make(map[string]*template.Template)
	//widgts, header, footer, sidebar, etc.
	var err error
	var bases []string
	var includes []string
	if r.Options.Includes != "" {
		includes, err = filepath.Glob(r.Options.Includes + "/*" + r.Options.Ext)
		if err != nil {
			logger.Fatal(err.Error())
		}
	}
	bases, err = filepath.Glob(r.Options.Templates + "/*" + r.Options.Ext)
	if err != nil {
		logger.Fatal(err.Error())
	}

	baseTmpl := template.New("").Funcs(r.Options.Funcs)
	if len(r.Options.Delims) >= 2 {
		baseTmpl.Delims(r.Options.Delims[0], r.Options.Delims[1])
	}

	// parse partials (glob)
	if len(includes) > 0 {
		baseTmpl = template.Must(baseTmpl.ParseFiles(includes...))
	}

	for _, templateFile := range bases {
		fileName := filepath.Base(templateFile)
		// set template name
		tmpl := template.Must(baseTmpl.Clone())
		tmpl = tmpl.New(fileName)
		// parse child template
		tmpl = template.Must(tmpl.ParseFiles(templateFile))
		templateSet[fileName] = tmpl
	}

	r.Templates.set = templateSet
}
