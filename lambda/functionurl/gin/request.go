package ginadapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type requestCtxKey struct{}

// Request returns the original events.LambdaFunctionURLRequest from context.
//
// If the gin handlers passed to StartBuffered or StartStream need access to the original request, it can be retrieved
// from context with this method.
func Request(ctx context.Context) *events.LambdaFunctionURLRequest {
	return ctx.Value(requestCtxKey{}).(*events.LambdaFunctionURLRequest)
}

type bufferedResponseCtxKey struct{}

// BufferedResponse returns the pending events.LambdaFunctionURLResponse from context.
//
// If the gin handlers passed to StartBuffered or StartStream need access to the pending response, it can be retrieved
// from context with this method.
func BufferedResponse(ctx context.Context) *events.LambdaFunctionURLResponse {
	return ctx.Value(bufferedResponseCtxKey{}).(*events.LambdaFunctionURLResponse)
}

type streamResponseCtxKey struct{}

// StreamResponse returns the pending events.LambdaFunctionURLStreamingResponse from context.
//
// If the gin handlers passed to StartBuffered or StartStream need access to the pending response, it can be retrieved
// from context with this method.
func StreamResponse(ctx context.Context) *events.LambdaFunctionURLStreamingResponse {
	return ctx.Value(streamResponseCtxKey{}).(*events.LambdaFunctionURLStreamingResponse)
}

func toHTTPRequest(req events.LambdaFunctionURLRequest) (r *http.Request, err error) {
	// http.NewRequest requires method, path, and request body.
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
			return nil, fmt.Errorf("decde base64-encoded request body error: %w", err)
		} else {
			body = bytes.NewReader(data)
		}
	}

	if r, err = http.NewRequest(method, path, body); err != nil {
		return nil, fmt.Errorf("create HTTP request error: %w", err)
	}

	// fill out more information from the request if possible.
	// https://docs.aws.amazon.com/lambda/latest/dg/urls-invocation.html#urls-payloads
	for _, v := range req.Cookies {
		r.Header.Add("Cookie", v)
	}
	// request header can show up for the same key multiple times with the values split by ",".
	for k, values := range req.Headers {
		k = http.CanonicalHeaderKey(k)
		for _, v := range strings.Split(values, ",") {
			r.Header.Add(k, v)
		}
	}

	r.Host, r.RemoteAddr, r.RequestURI = req.RequestContext.DomainName, req.RequestContext.HTTP.SourceIP, r.URL.RequestURI()

	return
}
