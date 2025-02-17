package ginadapter

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// StartStream starts the Lambda loop in STREAM_RESPONSE mode with the given Gin engine.
//
// Because StartStream uses a custom runtime, it does not accept lambda.Option settings. If you have such need, send me
// a PR.
func StartStream(r *gin.Engine) {
	baseURL, ok := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	if !ok {
		log.Fatalf("missing AWS_LAMBDA_RUNTIME_API from environment variables")
	} else if baseURL == "" {
		log.Fatalf("empty AWS_LAMBDA_RUNTIME_API from environment variables")
	}
	baseURL = "http://" + baseURL + "/2018-06-01/runtime/invocation"
}

// streamRuntime implements https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html#runtimes-custom-response-streaming.
type streamRuntime struct {
	client *http.Client
	buf    *bytes.Buffer
}

func startStreamRuntime() {

}
