package ginadapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/gin-gonic/gin"
)

// StartStream starts the Lambda loop in STREAM_RESPONSE mode with the given Gin engine.
//
// Because StartStream uses a custom runtime, it does not accept lambda.Option settings. If you have such need, send me
// a PR.
func StartStream(r *gin.Engine) {
	for req := range runtimeInvocations() {
		r.ServeHTTP(req, req.httpRequest)
	}
}

const (
	// https://docs.aws.amazon.com/lambda/latest/dg/runtimes-api.html#runtimes-api-next
	runtimeHeaderAWSRequestID       = "Lambda-Runtime-Aws-Request-Id"
	runtimeHeaderDeadlineMS         = "Lambda-Runtime-Deadline-Ms"
	runtimeHeaderInvokedFunctionARN = "Lambda-Runtime-Invoked-Function-Arn"
	runtimeHeaderCognitoIdentity    = "Lambda-Runtime-Cognito-Identity"
	runtimeHeaderClientContext      = "Lambda-Runtime-Client-Context"
)

// runtimeInvocations returns the next runtimeInvocations as an iterator (available since go1.23).
//
// Any runtime error will immediately stop execution with a log.Fatal for simplicity.
func runtimeInvocations() iter.Seq[*invocation] {
	endpoint, ok := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	if !ok {
		log.Fatalf("missing AWS_LAMBDA_RUNTIME_API from environment variables")
	}
	if endpoint == "" {
		log.Fatalf("empty AWS_LAMBDA_RUNTIME_API from environment variables")
	}

	baseURL := "http://" + endpoint + "/2018-06-01/runtime/invocation/"
	nextURL := baseURL + "next"

	return func(yield func(*invocation) bool) {
		resp, err := http.Get(nextURL)
		if err != nil {
			log.Fatalf("get invocation error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("get invocation returns non-OK status: %d (%s)", resp.StatusCode, resp.Status)
		}

		// the response header contains the deadline that can be used to create a cancellable context.
		deadlineMs, err := strconv.ParseInt(resp.Header.Get(runtimeHeaderDeadlineMS), 10, 64)
		if err != nil {
			log.Fatalf("parse deadline (%s) error: %v", resp.Header.Get(runtimeHeaderDeadlineMS), err)
		}
		ctx, cancel := context.WithDeadline(context.Background(), time.UnixMilli(deadlineMs))
		defer cancel()

		// we'll attach the LambdaContext for callers as well.
		lc := &lambdacontext.LambdaContext{
			AwsRequestID:       resp.Header.Get(runtimeHeaderAWSRequestID),
			InvokedFunctionArn: resp.Header.Get(runtimeHeaderInvokedFunctionARN),
		}
		if v := resp.Header.Get(runtimeHeaderClientContext); v != "" {
			if err = json.Unmarshal([]byte(v), &lc.ClientContext); err != nil {
				log.Fatalf("unmarshal client context error: %v", err)
			}
		}
		if v := resp.Header.Get(runtimeHeaderCognitoIdentity); v != "" {
			if err = json.Unmarshal([]byte(v), &lc.Identity); err != nil {
				log.Fatalf("unmarshal cognito idenity error: %v", err)
			}
		}
		ctx = lambdacontext.NewContext(ctx, lc)

		// decode JSON response body as LambdaFunctionURLRequest then convert to http.Request.
		request := &events.LambdaFunctionURLRequest{}
		err = json.NewDecoder(resp.Body).Decode(request)
		_ = resp.Body.Close()
		if err != nil {
			log.Fatalf("unmarshal invocation response body error: %v", err)
		}

		httpRequest, err := toHTTPRequest(request)
		if err != nil {
			log.Fatalf("create HTTP request from invocation response body error: %v", err)
		}

		httpRequest = httpRequest.WithContext(ctx)

		if !yield(&invocation{
			id:               lc.AwsRequestID,
			httpRequest:      httpRequest,
			ctx:              ctx,
			responseEndpoint: baseURL + lc.AwsRequestID + "/response",
			errEndpoint:      baseURL + lc.AwsRequestID + "/error",
			header:           make(http.Header),
		}) {
			cancel()
			return
		}
	}
}

// invocation is a single invocation from /runtime/invocation/next for STREAM_RESPONSE Function URL payloads.
type invocation struct {
	id               string
	httpRequest      *http.Request
	body             io.ReadCloser
	ctx              context.Context
	responseEndpoint string
	errEndpoint      string
	header           http.Header
	once             sync.Once
}

var _ http.ResponseWriter = &invocation{}
var _ http.ResponseWriter = (*invocation)(nil)
var _ http.Flusher = &invocation{}
var _ http.Flusher = (*invocation)(nil)

func (i *invocation) Header() http.Header {
	return i.header
}

func (i *invocation) Write(p []byte) (int, error) {
	i.once.Do(func() {
		i.WriteHeader(http.StatusOK)
	})

	//TODO implement me.
	panic("implement me")
}

func (i *invocation) WriteHeader(statusCode int) {
	i.once.Do(func() {
		i.WriteHeader(statusCode)
	})
}

func (i *invocation) Flush() {
	if err := i.body.Close(); err != nil {
		log.Fatalf("close connection to runtime error: %v", err)
	}
}

// writeHeader is not explicitly described in any documentation wrt streaming response.
// has to figure this out from https://github.com/aws/aws-lambda-go/blob/main/events/lambda_function_urls.go
// (events.LambdaFunctionURLStreamingResponse).
func (i *invocation) writeHeader(dst io.Writer, statusCode int) {
	cookies := make([]string, 0)
	headers := make(map[string]string)
	for k, vs := range i.header {
		if strings.EqualFold("Set-Cookie", k) {
			cookies = append(cookies, vs...)
		} else {
			headers[k] = strings.Join(vs, ",")
		}
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(struct {
		StatusCode int               `json:"statusCode"`
		Cookies    []string          `json:"cookies,omitempty"`
		Headers    map[string]string `json:"headers,omitempty"`
	}{
		StatusCode: statusCode,
		Cookies:    cookies,
		Headers:    headers,
	}); err != nil {
		log.Fatalf("marshal prelude error: %v", err)
	}
	_, _ = dst.Write(make([]byte, 8))

	req, err := http.NewRequest(http.MethodPost, i.responseEndpoint, body)
	if err != nil {
		log.Fatalf("create HTTP request error: %v", err)
	}

	req.Header = i.header
	req.Header.Add("Trailer", "Lambda-Runtime-Function-Error-Type")
	req.Header.Add("Trailer", "Lambda-Runtime-Function-Error-Body")
	req.Header.Set("Lambda-Runtime-Function-Response-Mode", "streaming")
	req.Header.Set("Transfer-Encoding", "chunked")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("post initial payload error: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		log.Fatalf("post initial payload returns non-Accept status: %d (%s)", resp.StatusCode, resp.Status)
	}

	// very important that we DO NOT close the body here. Flush will do that.
	i.body = resp.Body
}

func toHTTPRequest(req *events.LambdaFunctionURLRequest) (r *http.Request, err error) {
	// http.NewRequest requires method, path, and invocation body.
	reqCtx := req.RequestContext
	method := strings.ToUpper(reqCtx.HTTP.Method)

	// the path should also contain the query string if present.
	path := req.RawPath
	if path == "" {
		path = reqCtx.HTTP.Path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if req.RawQueryString != "" {
		path += "?" + req.RawQueryString
	} else if len(req.QueryStringParameters) > 0 {
		values := url.Values{}
		for k, v := range req.QueryStringParameters {
			values.Add(k, v)
		}
		path += "?" + values.Encode()
	}

	// the body may be base64-encoded in which case we'll decode it first.
	var body io.Reader = strings.NewReader(req.Body)
	if req.IsBase64Encoded {
		if data, err := base64.StdEncoding.DecodeString(req.Body); err != nil {
			return nil, fmt.Errorf("decde base64-encoded invocation body error: %w", err)
		} else {
			body = bytes.NewReader(data)
		}
	}

	if r, err = http.NewRequest(method, path, body); err != nil {
		return nil, fmt.Errorf("create HTTP invocation error: %w", err)
	}

	// fill out more information from the invocation if possible.
	// https://docs.aws.amazon.com/lambda/latest/dg/urls-invocation.html#urls-payloads
	for _, v := range req.Cookies {
		r.Header.Add("Cookie", v)
	}
	// invocation header can show up for the same key multiple times with the values split by ",".
	for k, values := range req.Headers {
		k = http.CanonicalHeaderKey(k)
		for _, v := range strings.Split(values, ",") {
			r.Header.Add(k, v)
		}
	}

	r.Host, r.RemoteAddr, r.RequestURI = req.RequestContext.DomainName, req.RequestContext.HTTP.SourceIP, r.URL.RequestURI()

	return
}
