package ginadapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
)

// StartBuffered starts the Lambda loop in BUFFERED mode with the given Gin engine.
func StartBuffered(r *gin.Engine, options ...lambda.Option) {
	lambda.StartHandlerFunc(func(ctx context.Context, req events.LambdaFunctionURLRequest) (res events.LambdaFunctionURLResponse, err error) {
		httpRequest, err := toHTTPRequest(req)
		if err != nil {
			res.StatusCode = http.StatusBadGateway
			return res, err
		}

		w := &bufferedResponseWriter{
			statusCode: 0,
			header:     make(http.Header),
			buf:        &bytes.Buffer{},
		}

		// this is really where the magic happens. because gin.Engine implements http.Handler interface, we can
		// use it like this.
		r.ServeHTTP(w, httpRequest)

		res.StatusCode = w.statusCode

		// cookies and headers come from the same w.header.
		res.Cookies = make([]string, 0)
		res.Headers = make(map[string]string)
		for k, vs := range w.header {
			if strings.EqualFold("Set-Cookie", k) {
				res.Cookies = append(res.Cookies, vs...)
			} else {
				res.Headers[k] = strings.Join(vs, ",")
			}
		}

		// if we can detect that the body is a valid UTF8-string then don't base64 encode it.
		if b := w.buf.Bytes(); utf8.Valid(b) {
			res.Body = string(b)
		} else {
			res.Body = base64.StdEncoding.EncodeToString(b)
			res.IsBase64Encoded = true
		}

		return
	}, options...)
}

// bufferedResponseWriter implements http.ResponseWriter to serve as the adapter between gin and Lambda.
type bufferedResponseWriter struct {
	statusCode int
	header     http.Header
	buf        *bytes.Buffer
}

var _ http.ResponseWriter = &bufferedResponseWriter{}
var _ http.ResponseWriter = (*bufferedResponseWriter)(nil)

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) Write(bytes []byte) (int, error) {
	return w.buf.Write(bytes)
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
