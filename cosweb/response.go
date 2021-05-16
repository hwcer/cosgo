package cosweb

import (
	"bufio"
	"github.com/hwcer/cosgo/logger"
	"net"
	"net/http"
)

func (c *Context) Header() http.Header {
	return c.Response.Header()
}

// Write writes the store to the connection as part of an HTTP reply.
func (c *Context) Write(b []byte) (n int, err error) {
	if !c.committed {
		c.WriteHeader(0)
	}
	n, err = c.Response.Write(b)
	//c.contentSize += int64(n)
	return
}

// Status sends an HTTP Response header with status code. If Status is
// not called explicitly, the first call to Write will trigger an implicit
// Status(http.StatusOK). Thus explicit calls to Status are mainly
// used to send error codes.
func (c *Context) WriteHeader(code int) {
	if c.committed {
		logger.Error("WriteHeader but response already committed")
		return
	}
	if code == 0 {
		code = http.StatusOK
	}
	c.Response.WriteHeader(code)
	c.committed = true
}

// Flush implements the http.Flusher interface to allow an HTTP Handler to flush
// buffered store to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (c *Context) Flush() {
	c.Response.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP Handler to
// take over the connection.
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (c *Context) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return c.Response.(http.Hijacker).Hijack()
}
