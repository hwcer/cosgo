package cosweb

import (
	"bufio"
	"net"
	"net/http"
)

func (c *Context) Header() http.Header {
	return c.Response.Header()
}

// Write writes the store to the connection as part of an HTTP reply.
func (c *Context) Write(b []byte) (n int, err error) {
	//c.state.Push(netStateTypeWriteComplete)
	//c.WriteHeader(0)
	n, err = c.Response.Write(b)
	//c.contentSize += int64(n)
	return
}

// Status sends an HTTP Response header with status code. If Status is
// not called explicitly, the first call to Write will trigger an implicit
// Status(http.StatusOK). Thus explicit calls to Status are mainly
// used to send error codes.
func (c *Context) WriteHeader(code int) {
	c.Response.WriteHeader(code)
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered store to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (c *Context) Flush() {
	c.Response.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (c *Context) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return c.Response.(http.Hijacker).Hijack()
}
