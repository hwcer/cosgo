package render

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
)

// Render provides functions for easily writing HTML templates & JSON out to a HTTP Response.
type Render struct {
	Options   *Options
	Templates *templates
}

// Options holds the configuration Options for a Render
type Options struct {
	// With Debug set to true, templates will be recompiled before every render call.
	Debug bool

	// The glob string to your templates
	TemplatesGlob string

	// The Glob string for additional templates
	PartialsGlob string

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
	if r.Options.TemplatesGlob == "" {
		return
	}

	// replace existing templates.
	// NOTE: this is unsafe, but Debug should really not be true in production environments.
	templateSet := make(map[string]*template.Template)

	files, err := filepath.Glob(r.Options.TemplatesGlob)
	if err != nil {
		panic(err)
	}

	baseTmpl := template.New("").Funcs(r.Options.Funcs)
	if len(r.Options.Delims) >= 2 {
		baseTmpl.Delims(r.Options.Delims[0], r.Options.Delims[1])
	}

	// parse partials (glob)
	if r.Options.PartialsGlob != "" {
		baseTmpl = template.Must(baseTmpl.ParseGlob(r.Options.PartialsGlob))
	}

	for _, templateFile := range files {
		fileName := filepath.Base(templateFile)
		layout := getLayoutForTemplate(templateFile)

		// set template name
		name := fileName
		if layout != "" {
			name = filepath.Base(layout)
		}

		tmpl := template.Must(baseTmpl.Clone())
		tmpl = tmpl.New(name)

		// parse master template
		if layout != "" {
			layoutFile := filepath.Join(filepath.Dir(templateFile), layout)
			tmpl = template.Must(tmpl.ParseFiles(layoutFile))
		}

		// parse child template
		tmpl = template.Must(tmpl.ParseFiles(templateFile))

		templateSet[fileName] = tmpl
	}

	r.Templates.set = templateSet
}
