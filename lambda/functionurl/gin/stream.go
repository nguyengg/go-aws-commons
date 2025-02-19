package ginadapter

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/metrics"
	"github.com/rs/zerolog"
)

// StartStream starts the Lambda loop in STREAM_RESPONSE mode with the given Gin engine.
func StartStream(r *gin.Engine, options ...lambda.Option) {
	r.Use(fault)

	// because gin.Engine implements http.Handler interface, lambdaurl already provides this adapter for me.
	lambdaurl.Start(handler{r}, options...)
}

// fault is a gin middleware that emits fault counter.
func fault(c *gin.Context) {
	c.Next()

	if err := c.Errors.Last(); err != nil {
		// we can't use metrics.Ctx(c) because we can't rely on user enabling gin.Engine.ContextWithFallback.
		// we technically can do that for user, but that's more dangerous than prepending our own middleware.
		metrics.Ctx(c.Request.Context()).Faulted()
	}
}

// handler wraps the gin.Engine to provide metrics.
type handler struct {
	r *gin.Engine
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m := metrics.New()

	if lc, ok := lambdacontext.FromContext(ctx); ok {
		m.SetProperty("awsRequestID", lc.AwsRequestID)
	}

	panicked := true
	defer func() {
		if panicked {
			m.Panicked()
		}

		m.Log(zerolog.Ctx(ctx))
	}()

	*r = *r.WithContext(metrics.WithContext(ctx, m))

	h.r.ServeHTTP(&writer{w, m}, r)

	panicked = false
}

// writer wraps the http.ResponseWriter to update the metrics instance's status code.
type writer struct {
	http.ResponseWriter
	m metrics.Metrics
}

func (w *writer) WriteHeader(statusCode int) {
	w.m.SetStatusCode(statusCode)
	w.ResponseWriter.WriteHeader(statusCode)
}
