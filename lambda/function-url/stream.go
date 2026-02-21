package functionurl

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/nguyengg/go-aws-commons/metrics"
	"github.com/rotisserie/eris"
)

// StartStream starts the Lambda loop in STREAM_RESPONSE mode with the given handler.
//
// gin.Engine satisfies http.Handler so you should pass one here.
func StartStream(r http.Handler, options ...lambda.Option) {
	lambdaurl.Start(wrapper{r}, options...)
}

// handler wraps the gin.Engine to provide metrics.
type wrapper struct {
	http.Handler
}

func (h wrapper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx, m := metrics.NewWithContext(req.Context())

	// TODO make this configurable.
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		log.SetPrefix(lc.AwsRequestID + " ")
		slog.SetDefault(slog.With("awsRequestId", lc.AwsRequestID))
		m.String("awsRequestId", lc.AwsRequestID)
	}

	w := &writer{rw, m, 0}

	defer func() {
		if r := recover(); r != nil {
			m.Panicked()

			switch v := r.(type) {
			case error:
				m.Error(v)
			default:
				m.Error(eris.Wrapf(fmt.Errorf("%+v", v), "recover non-error %T: %#v", v, v))
			}
		}

		_ = m.CloseContext(ctx)
	}()

	h.Handler.ServeHTTP(w, req.WithContext(ctx))

	m.Int64("size", w.size)
}

// writer wraps the http.ResponseWriter to update the metrics instance's status code.
type writer struct {
	http.ResponseWriter
	*metrics.Metrics
	size int64
}

func (w *writer) Write(p []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *writer) WriteHeader(statusCode int) {
	w.Int64("status", int64(statusCode))
	w.ResponseWriter.WriteHeader(statusCode)
}

var _ http.Flusher = &writer{}
var _ http.Flusher = (*writer)(nil)

func (w *writer) Flush() {
	// gin.Engine also implements http.Flusher.
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
