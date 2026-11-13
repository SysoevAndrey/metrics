package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	textStr = "Plain Text"
	htmlStr = "<html><body>Hello, world</body></html>"
	jsonStr = `{"value": 2}`
)

func TestWithCompression(t *testing.T) {
	type request struct {
		method          string
		path            string
		contentType     string
		contentEncoding string
		acceptEncoding  string
		body            io.Reader
	}

	type response struct {
		statusCode      int
		contentEncoding string
		body            io.Reader
	}

	tests := []struct {
		name     string
		request  request
		response response
	}{
		{
			name: "compress html",
			request: request{
				method:         http.MethodGet,
				path:           "/html",
				acceptEncoding: "gzip",
			},
			response: response{
				statusCode:      200,
				contentEncoding: "gzip",
				body:            bytes.NewBufferString(htmlStr),
			},
		},
		{
			name: "compress json",
			request: request{
				method:         http.MethodGet,
				path:           "/json",
				acceptEncoding: "gzip",
			},
			response: response{
				statusCode:      200,
				contentEncoding: "gzip",
				body:            bytes.NewBufferString(jsonStr),
			},
		},
		{
			name: "doesn't compress text",
			request: request{
				method: http.MethodGet,
				path:   "/text",
			},
			response: response{
				statusCode: 200,
				body:       bytes.NewBufferString(textStr),
			},
		},
		{
			name: "doesn't compress html",
			request: request{
				method: http.MethodGet,
				path:   "/html",
			},
			response: response{
				statusCode: 200,
				body:       bytes.NewBufferString(htmlStr),
			},
		},
		{
			name: "doesn't compress json",
			request: request{
				method: http.MethodGet,
				path:   "/json",
			},
			response: response{
				statusCode: 200,
				body:       bytes.NewBufferString(jsonStr),
			},
		},
		{
			name: "decompress json, returns compressed json",
			request: request{
				method:          http.MethodPost,
				path:            "/json",
				contentType:     "application/json",
				contentEncoding: "gzip",
				acceptEncoding:  "gzip",
				body:            bytes.NewBufferString(jsonStr),
			},
			response: response{
				statusCode:      200,
				contentEncoding: "gzip",
				body:            bytes.NewBufferString(jsonStr),
			},
		},
		{
			name: "decompress json, returns json",
			request: request{
				method:          http.MethodPost,
				path:            "/json",
				contentType:     "application/json",
				contentEncoding: "gzip",
				body:            bytes.NewBufferString(jsonStr),
			},
			response: response{
				statusCode: 200,
				body:       bytes.NewBufferString(jsonStr),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			if tt.request.contentEncoding == "gzip" && tt.request.body != nil {
				zw := gzip.NewWriter(buf)
				_, err := io.Copy(zw, tt.request.body)
				require.NoError(t, err)
				err = zw.Close()
				require.NoError(t, err)
			} else if tt.request.body != nil {
				_, err := io.Copy(buf, tt.request.body)
				require.NoError(t, err)
			}

			request := httptest.NewRequest(tt.request.method, tt.request.path, buf)
			if len(tt.request.contentType) > 0 {
				request.Header.Set("Content-Type", tt.request.contentType)
			}
			if len(tt.request.contentEncoding) > 0 {
				request.Header.Set("Content-Encoding", tt.request.contentEncoding)
			}
			if len(tt.request.acceptEncoding) > 0 {
				request.Header.Set("Accept-Encoding", tt.request.acceptEncoding)
			}

			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Use(WithCompression)
			r.Get("/text", textHandler)
			r.Get("/html", htmlHandler)
			r.Get("/json", jsonHandler)
			r.Post("/json", jsonHandler)
			r.ServeHTTP(w, request)

			result := w.Result()

			assert.Equal(t, tt.response.statusCode, result.StatusCode)

			defer result.Body.Close()
			bodyRaw := []byte{}

			if tt.response.contentEncoding == "gzip" {
				contentEncoding := result.Header.Get("Content-Encoding")
				assert.Equal(t, tt.response.contentEncoding, contentEncoding)

				zr, err := gzip.NewReader(result.Body)
				require.NoError(t, err)
				body, err := io.ReadAll(zr)
				require.NoError(t, err)
				bodyRaw = body
			}

			if len(tt.request.acceptEncoding) == 0 {
				contentEncoding := result.Header.Get("Content-Encoding")
				assert.Empty(t, contentEncoding)

				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				bodyRaw = body
			}

			expectedRaw, err := io.ReadAll(tt.response.body)
			require.NoError(t, err)
			assert.Equal(t, expectedRaw, bodyRaw)
		})
	}
}

func textHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(textStr))
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlStr))
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonStr))
}
