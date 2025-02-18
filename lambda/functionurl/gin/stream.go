package ginadapter

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/gin-gonic/gin"
)

// StartStream starts the Lambda loop in STREAM_RESPONSE mode with the given Gin engine.
func StartStream(r *gin.Engine, options ...lambda.Option) {
	// because gin.Engine implements http.Handler interface, lambdaurl already provides this adapter for me.
	lambdaurl.Start(r, options...)
}
