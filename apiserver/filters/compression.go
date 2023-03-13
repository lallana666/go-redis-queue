package filters

import (
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Compressor is an interface to compression writers
type Compressor interface {
	io.WriteCloser
	Flush() error
}

const (
	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerContentType     = "Content-Type"

	encodingGzip    = "gzip"
	encodingDeflate = "deflate"
)

// WithCompression wraps an http.Handler with the Compression Handler
func WithCompression(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantsCompression, encoding := wantsCompressedResponse(r)
		w.Header().Set("Vary", "Accept-Encoding")
		if wantsCompression {
			compressionWriter, err := NewCompressionResponseWriter(w, encoding)
			if err != nil {
				errorMsg := fmt.Sprintf("Internal Server Error: %#v", r.RequestURI)
				http.Error(w, errorMsg, http.StatusInternalServerError)
				return
			}
			compressionWriter.Header().Set("Content-Encoding", encoding)
			handler.ServeHTTP(compressionWriter, r)
			compressionWriter.(*compressionResponseWriter).Close()
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}

// wantsCompressedResponse reads the Accept-Encoding header to see if and which encoding is requested.
func wantsCompressedResponse(r *http.Request) (bool, string) {
	header := r.Header.Get(headerAcceptEncoding)
	gi := strings.Index(header, encodingGzip)
	zi := strings.Index(header, encodingDeflate)
	// use in order of appearance
	switch {
	case gi == -1:
		return zi != -1, encodingDeflate
	case zi == -1:
		return gi != -1, encodingGzip
	case gi < zi:
		return true, encodingGzip
	default:
		return true, encodingDeflate
	}
}

type compressionResponseWriter struct {
	writer     http.ResponseWriter
	compressor Compressor
	encoding   string
}

// NewCompressionResponseWriter returns wraps w with a compression ResponseWriter, using the given encoding
func NewCompressionResponseWriter(w http.ResponseWriter, encoding string) (http.ResponseWriter, error) {
	var compressor Compressor
	switch encoding {
	case encodingGzip:
		compressor = gzip.NewWriter(w)
	case encodingDeflate:
		compressor = zlib.NewWriter(w)
	default:
		return nil, fmt.Errorf("%s is not a supported encoding type", encoding)
	}
	return &compressionResponseWriter{
		writer:     w,
		compressor: compressor,
		encoding:   encoding,
	}, nil
}

// compressionResponseWriter implements http.ResponseWriter Interface
var _ http.ResponseWriter = &compressionResponseWriter{}

func (c *compressionResponseWriter) Header() http.Header {
	return c.writer.Header()
}

// compress data according to compression method
func (c *compressionResponseWriter) Write(b []byte) (int, error) {
	if c.compressorClosed() {
		return -1, errors.New("compressing error: tried to write data using closed compressor")
	}
	c.Header().Set(headerContentEncoding, c.encoding)
	if len(c.Header().Get(headerContentType)) == 0 {
		c.Header().Set(headerContentType, http.DetectContentType(b))
	}
	defer c.compressor.Flush()
	return c.compressor.Write(b)
}

func (c *compressionResponseWriter) WriteHeader(status int) {
	c.writer.WriteHeader(status)
}

// CloseNotify is part of http.CloseNotifier interface
func (c *compressionResponseWriter) CloseNotify() <-chan bool {
	return c.writer.(http.CloseNotifier).CloseNotify()
}

// Close the underlying compressor
func (c *compressionResponseWriter) Close() error {
	if c.compressorClosed() {
		return errors.New("Compressing error: tried to close already closed compressor")
	}

	c.compressor.Close()
	c.compressor = nil
	return nil
}

func (c *compressionResponseWriter) Flush() {
	if c.compressorClosed() {
		return
	}
	c.compressor.Flush()
}

func (c *compressionResponseWriter) compressorClosed() bool {
	return nil == c.compressor
}
