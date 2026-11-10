package middleware

import (
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"strings"
)

var allowedContentTypes map[string]struct{} = map[string]struct{}{
	"application/json": {},
	"text/html":        {},
}

type compressWriter struct {
	w              http.ResponseWriter
	zw             *gzip.Writer
	shouldCompress *bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:              w,
		zw:             gzip.NewWriter(w),
		shouldCompress: nil,
	}
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressWriter) Write(p []byte) (int, error) {
	cw.setCompressionIfNeeded()

	if !*cw.shouldCompress {
		return cw.w.Write(p)
	}
	return cw.zw.Write(p)
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	cw.setCompressionIfNeeded()
	cw.w.WriteHeader(statusCode)
}

func (cw *compressWriter) Close() error {
	if cw.shouldCompress != nil && *cw.shouldCompress {
		return cw.zw.Close()
	}
	return nil
}

func (cw *compressWriter) setCompressionIfNeeded() {
	if cw.shouldCompress != nil {
		return
	}

	contentType := cw.w.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		cw.shouldCompress = new(false)
		return
	}

	_, allowed := allowedContentTypes[mediaType]
	if allowed {
		cw.shouldCompress = new(true)
		cw.w.Header().Set("Content-Encoding", "gzip")
	} else {
		cw.shouldCompress = new(false)
	}
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (cr *compressReader) Read(p []byte) (int, error) {
	return cr.zr.Read(p)
}

func (cr *compressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		return err
	}
	return cr.zr.Close()
}

func WithCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			defer cw.Close()
			ow = cw
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer cr.Close()
			r.Body = cr
		}

		h.ServeHTTP(ow, r)
	})
}
