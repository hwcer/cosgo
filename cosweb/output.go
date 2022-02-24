package cosweb

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (c *Context) writeContentType(contentType ContentType) {
	header := c.Header()
	header.Set(HeaderContentType, GetContentTypeCharset(contentType))
}

func (c *Context) contentDisposition(file, name, dispositionType string) error {
	header := c.Header()
	header.Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

func (c *Context) Bytes(contentType ContentType, b []byte) (err error) {
	c.writeContentType(contentType)
	_, err = c.Write(b)
	return
}
func (c *Context) Render(name string, data interface{}) (err error) {
	if c.engine.Render == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.engine.Render.Render(buf, name, data); err != nil {
		return
	}
	return c.Bytes(ContentTypeTextHTML, buf.Bytes())
}

func (c *Context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(c, c.Request, fi.Name(), fi.ModTime(), f)
	return
}

func (c *Context) Stream(contentType ContentType, r io.Reader) (err error) {
	c.writeContentType(contentType)
	_, err = io.Copy(c.Response, r)
	return
}

//Inline 最终走File
func (c *Context) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

//Attachment 最终走File
func (c *Context) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *Context) Redirect(url string) error {
	c.Response.Header().Set(HeaderLocation, url)
	c.WriteHeader(http.StatusFound)
	return nil
}

func (c *Context) XML(i interface{}, indent string) (err error) {
	data, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	c.Bytes(ContentTypeApplicationXML, data)
	return
}

func (c *Context) HTML(html string) (err error) {
	return c.Bytes(ContentTypeTextHTML, []byte(html))
}

func (c *Context) String(s string) (err error) {
	return c.Bytes(ContentTypeTextPlain, []byte(s))
}

func (c *Context) JSON(i interface{}) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.Bytes(ContentTypeApplicationJSON, data)
}

func (c *Context) JSONP(callback string, i interface{}) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	data = bytes.Join([][]byte{[]byte(callback), []byte("("), data, []byte(")")}, []byte{})
	return c.Bytes(ContentTypeApplicationJS, data)
}
