package ginadapter

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/metrics"
)

// StartStream starts the Lambda loop in STREAM_RESPONSE mode with the given Gin engine.
func StartStream(r *gin.Engine, options ...lambda.Option) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{ReplaceAttr: metrics.ReplaceAttr()})))

	r.Use(fault)

	// because gin.Engine implements http.Handler interface, lambdaurl already provides this adapter for me.
	lambdaurl.Start(handler{r}, options...)
}

// fault is a gin middleware that emits fault counter.
func fault(c *gin.Context) {
	c.Next()

	if err := c.Errors.Last(); err != nil {
		// we can't use metrics.Get(c) because we can't rely on user enabling gin.Engine.ContextWithFallback.
		// we technically can do that for user, but that's more dangerous than prepending our own middleware.
		metrics.Get(c.Request.Context()).Faulted()
	}
}

// handler wraps the gin.Engine to provide metrics.
type handler struct {
	r *gin.Engine
}

func (h handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx, m := metrics.NewWithContext(req.Context())

	if lc, ok := lambdacontext.FromContext(ctx); ok {
		log.SetPrefix(lc.AwsRequestID + " ")
		slog.SetDefault(slog.With("awsRequestId", lc.AwsRequestID))
		m.String("awsRequestId", lc.AwsRequestID)
	}

	w := &writer{rw, m, nil}

	defer func() {
		if r := recover(); r != nil {
			m.Panicked()
			m.Any("error", r)
		} else if w.err != nil {
			m.Faulted()
			m.Any("error", w.err)
		}

		_ = m.CloseContext(ctx)
	}()

	h.r.ServeHTTP(w, req.WithContext(ctx))
}

// writer wraps the http.ResponseWriter to update the metrics instance's status code.
type writer struct {
	http.ResponseWriter
	*metrics.Metrics

	err error
}

func (w *writer) Write(p []byte) (n int, _ error) {
	n, w.err = w.ResponseWriter.Write(p)
	return n, w.err
}

func (w *writer) WriteHeader(statusCode int) {
	w.Int64("status", int64(statusCode))
	w.ResponseWriter.WriteHeader(statusCode)
}

var _ http.Flusher = &writer{}
var _ http.Flusher = (*writer)(nil)

func (w *writer) Flush() {
	// the gin's writer implements flush already.
	w.ResponseWriter.(http.Flusher).Flush()
}
