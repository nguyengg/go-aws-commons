package ginadapter

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func Test_toHTTPRequest(t *testing.T) {
	type wantR struct {
		// pick attributes from http.Request to compare
		Method        string
		URL           string
		Header        map[string][]string
		ContentLength int64
		Host          string
		RemoteAddr    string
		RequestURI    string
	}
	tests := []struct {
		name    string
		req     events.LambdaFunctionURLRequest
		want    wantR
		wantErr bool
	}{
		{
			name: "success",
			req: events.LambdaFunctionURLRequest{
				RawPath:        "/hello-world",
				RawQueryString: "hello=world",
				Cookies:        []string{"sid=12345"},
				Headers: map[string]string{
					"x-amz-request-id": "12345",       // should be made "canonical"
					"Test":             "hello,world", // example of split cookie
				},
				QueryStringParameters: nil, // RawQueryString is used.
				RequestContext: events.LambdaFunctionURLRequestContext{
					DomainName: "1234.lambda-url.us-west-2.on.aws",
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method:   "get",
						Path:     "", // RawPath is used.
						SourceIP: "127.0.0.1",
					},
				},
				Body:            "",
				IsBase64Encoded: false,
			},
			want: wantR{
				Method: http.MethodGet,
				URL:    "/hello-world?hello=world",
				Header: map[string][]string{
					"Cookie":           {"sid=12345"},
					"X-Amz-Request-Id": {"12345"},
					"Test":             {"hello", "world"},
				},
				ContentLength: 0,
				Host:          "1234.lambda-url.us-west-2.on.aws",
				RemoteAddr:    "127.0.0.1",
				RequestURI:    "/hello-world?hello=world",
			},
		},
		{
			name: "alternative",
			req: events.LambdaFunctionURLRequest{
				RawPath:        "", // RequestContext.HTTP.Path is used.
				RawQueryString: "", // QueryStringParameters is used.
				QueryStringParameters: map[string]string{
					"hello": "world",
				},
				RequestContext: events.LambdaFunctionURLRequestContext{
					DomainName: "1234.lambda-url.us-west-2.on.aws",
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method:   "post",
						Path:     "/hello/world",
						SourceIP: "127.0.0.1",
					},
				},
				Body:            "i'm a teapot",
				IsBase64Encoded: false,
			},
			want: wantR{
				Method:        http.MethodPost,
				URL:           "/hello/world?hello=world",
				Header:        map[string][]string{},
				ContentLength: int64(len("i'm a teapot")),
				Host:          "1234.lambda-url.us-west-2.on.aws",
				RemoteAddr:    "127.0.0.1",
				RequestURI:    "/hello/world?hello=world",
			},
		},
		{
			name: "base64-encoded body",
			req: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					DomainName: "1234.lambda-url.us-west-2.on.aws",
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "put",
					},
				},
				Body:            base64.StdEncoding.EncodeToString([]byte("i'm a teapot")),
				IsBase64Encoded: true,
			},
			want: wantR{
				Method:        http.MethodPut,
				URL:           "/",
				Header:        map[string][]string{},
				ContentLength: int64(len("i'm a teapot")),
				Host:          "1234.lambda-url.us-west-2.on.aws",
				RequestURI:    "/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toHTTPRequest(tt.req)
			if tt.wantErr {
				assert.Errorf(t, err, "toHTTPRequest() want error but none returned")
			} else {
				assert.NoErrorf(t, err, "toHTTPRequest() error = %v", err)
			}

			assert.Equal(t, tt.want, wantR{
				Method:        got.Method,
				URL:           got.URL.String(),
				Header:        got.Header,
				ContentLength: got.ContentLength,
				Host:          got.Host,
				RemoteAddr:    got.RemoteAddr,
				RequestURI:    got.RequestURI,
			})
		})
	}
}
