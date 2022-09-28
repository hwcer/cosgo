package binding

import (
	"errors"
	"io"
	"mime"
)

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              string = "application/json"
	MIMEHTML                     = "text/html"
	MIMEXML                      = "application/xml"
	MIMEXML2                     = "text/xml"
	MIMEPlain                    = "text/plain"
	MIMEPOSTForm                 = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm        = "multipart/form-data"
	MIMEPROTOBUF                 = "application/x-protobuf"
	MIMEMSGPACK                  = "application/x-msgpack"
	MIMEMSGPACK2                 = "application/msgpack"
	MIMEYAML                     = "application/x-yaml"
)

var bindingMap = make(map[string]Binding)

// Binding describes the interface which needs to be implemented for binding the
// data present in the request such as JSON request body, query parameters or
// the form POST.
type Binding interface {
	Name() string
	Bind(io.Reader, interface{}) error
	Unmarshal([]byte, interface{}) error
}

func Handle(name string) (b Binding) {
	ct, _, err := mime.ParseMediaType(name)
	if err == nil {
		b = bindingMap[ct]
	}
	return
}

func Register(name string, handle Binding) error {
	ct, _, err := mime.ParseMediaType(name)
	if err != nil {
		return err
	}
	bindingMap[ct] = handle
	return nil
}

func Bind(io io.Reader, i interface{}, name string) error {
	handle := Handle(name)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Bind(io, i)
}

func Unmarshal(b []byte, i interface{}, name string) error {
	handle := Handle(name)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Unmarshal(b, i)
}
