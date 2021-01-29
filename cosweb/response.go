package cosweb

import (
	"bufio"
	"cosgo/logger"
	"net"
	"net/http"
)

type (
	// Response wraps an http.ResponseWriter and implements its interface to be used
	// by an HTTP Handler to construct an HTTP Response.
	// See: https://golang.org/pkg/net/http/#ResponseWriter
	Response struct {
		engine         *Server
		Writer         http.ResponseWriter
		committed      bool
		contentSize    int64
		httpStatusCode int
	}
)

// NewResponse creates a new instance of Response.
func NewResponse(w http.ResponseWriter, e *Server) (r *Response) {
	return &Response{Writer: w, engine: e}
}

// Header returns the header map for the writer that will be sent by
// Status. Changing the header after a call to Status (or Write) has
// no effect unless the modified headers were declared as trailers by setting
// the "Trailer" header before the call to Status (see example)
// To suppress implicit Response headers, set their value to nil.
// Example: https://golang.org/pkg/net/http/#example_ResponseWriter_trailers
func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

// Status sends an HTTP Response header with status code. If Status is
// not called explicitly, the first call to Write will trigger an implicit
// Status(http.StatusOK). Thus explicit calls to Status are mainly
// used to send error codes.
func (r *Response) WriteHeader(code int) {
	if r.committed {
		logger.Error("WriteHeader but response already committed")
		return
	}
	if code > 0 {
		r.httpStatusCode = code
	}
	if r.httpStatusCode == 0 {
		r.httpStatusCode = http.StatusOK
	}
	r.Writer.WriteHeader(r.httpStatusCode)
	r.committed = true
}

func (r *Response) Status(code int) {
	if r.committed {
		logger.Error("set status but Response already committed")
		return
	}
	r.httpStatusCode = code
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *Response) Write(b []byte) (n int, err error) {
	if !r.committed {
		r.WriteHeader(r.httpStatusCode)
	}
	n, err = r.Writer.Write(b)
	r.contentSize += int64(n)
	return
}

// Flush implements the http.Flusher interface to allow an HTTP Handler to flush
// buffered data to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *Response) Flush() {
	r.Writer.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP Handler to
// take over the connection.
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.Writer.(http.Hijacker).Hijack()
}

func (r *Response) reset(w http.ResponseWriter) {
	r.Writer = w
	r.committed = false
	r.contentSize = 0
	r.httpStatusCode = 0
}
