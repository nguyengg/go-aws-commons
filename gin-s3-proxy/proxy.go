// Package proxy provides a simple Gin middleware to act as proxy to S3.
//
// Known limitation: a single PutObject is used to handle PUT and POST so be mindful of the 5GB maximum object size.
package proxy

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstransporthttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	copyheaders "github.com/nguyengg/go-aws-commons/gin-s3-proxy/copy-headers"
)

// S3APIClient extracts only the methods used by proxy to make unit testing easier.
type S3APIClient interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// New creates an S3 proxy middleware.
func New(client S3APIClient, getter BucketKeyGetter, opts ...Option) gin.HandlerFunc {
	cfg := &Config{bucketKeyGetter: getter}
	for _, opt := range opts {
		opt(cfg)
	}

	skip := cfg.Skip
	if skip == nil {
		skip = func(c *gin.Context) bool {
			return false
		}
	}

	emptyHandler := cfg.EmptyBucketOrKeyHandler
	if emptyHandler == nil {
		emptyHandler = func(c *gin.Context) {
			c.AbortWithStatus(http.StatusNotFound)
		}
	}

	return func(c *gin.Context) {
		if skip(c) {
			return
		}

		bucket, key, versionId := cfg.bucketKeyGetter(c)
		if bucket == "" || key == "" {
			emptyHandler(c)
			return
		}

		switch c.Request.Method {
		case http.MethodGet:
			input := copyheaders.ToGetObjectInput(c, &s3.GetObjectInput{Bucket: &bucket, Key: &key, VersionId: versionId})
			if cfg.ModifyGetObjectInput != nil {
				cfg.ModifyGetObjectInput(input)
			}

			output, err := client.GetObject(c, input)
			if err != nil {
				handleErr(c, err)
				return
			}

			copyheaders.FromGetObjectOutput(c, output)
			c.DataFromReader(200, aws.ToInt64(output.ContentLength), aws.ToString(output.ContentType), output.Body, nil)
			return

		case http.MethodHead:
			input := copyheaders.ToHeadObjectInput(c, &s3.HeadObjectInput{Bucket: &bucket, Key: &key, VersionId: versionId})
			if cfg.ModifyHeadObjectInput != nil {
				cfg.ModifyHeadObjectInput(input)
			}

			output, err := client.HeadObject(c, input)
			if err != nil {
				handleErr(c, err)
				return
			}

			copyheaders.FromHeadObjectOutput(c, output)
			c.Status(http.StatusOK)
			return

		case http.MethodDelete:
			input := copyheaders.ToDeleteObjectInput(c, &s3.DeleteObjectInput{Bucket: &bucket, Key: &key, VersionId: versionId})
			if cfg.ModifyDeleteObjectInput != nil {
				cfg.ModifyDeleteObjectInput(input)
			}

			if _, err := client.DeleteObject(c, input); err != nil {
				handleErr(c, err)
				return
			}

		case http.MethodPut, http.MethodPost:
			input := copyheaders.ToPutObjectInput(c, &s3.PutObjectInput{Bucket: &bucket, Key: &key, Body: c.Request.Body})
			if cfg.ModifyPutObjectInput != nil {
				cfg.ModifyPutObjectInput(input)
			}

			output, err := client.PutObject(c, input)
			if err != nil {
				handleErr(c, err)
			}

			copyheaders.FromPutObjectOutput(c, output)
			c.Status(http.StatusNoContent)
		}
	}
}

// Config customises the S3 proxy.
type Config struct {
	// Skip is used to abort or manually handle requests that the proxy should skip.
	//
	// By default, only GET, HEAD, DELETE, PUT, and POST methods are supported; any other methods immediately get a 404
	// regardless of whether the resource exists.
	Skip Skipper

	// EmptyBucketOrKeyHandler is called if either bucket or key cannot be retrieved from request.
	//
	// By default, the request is aborted with a 404.
	EmptyBucketOrKeyHandler func(c *gin.Context)

	// ModifyGetObjectInput allows additional modifications such as adding ExpectedBucketOwner to the input.
	ModifyGetObjectInput func(input *s3.GetObjectInput)
	// ModifyHeadObjectInput allows additional modifications such as adding ExpectedBucketOwner to the input.
	ModifyHeadObjectInput func(input *s3.HeadObjectInput)
	// ModifyDeleteObjectInput allows additional modifications such as adding ExpectedBucketOwner to the input.
	ModifyDeleteObjectInput func(input *s3.DeleteObjectInput)
	// ModifyPutObjectInput allows additional modifications such as adding ExpectedBucketOwner to the input.
	ModifyPutObjectInput func(input *s3.PutObjectInput)

	bucketKeyGetter BucketKeyGetter
}

// Option customises Config.
type Option func(cfg *Config)

// BucketKeyGetter defines how the bucket, key, and optional version Id are retrieved.
type BucketKeyGetter func(c *gin.Context) (bucket string, key string, versionId *string)

// WithBucketKeyFromParams will retrieve the bucket and key from [gin.Context.Param].
func WithBucketKeyFromParams(bucketParamName, keyParamName string) BucketKeyGetter {
	return func(c *gin.Context) (string, string, *string) {
		return c.Param(bucketParamName), c.Param(keyParamName), nil
	}
}

// WithBucketKeyAndVersionId will retrieve the bucket and key from [gin.Context.Param], optional version Id from
// [gin.Context.Query].
func WithBucketKeyAndVersionId(bucketParamName, keyParamName, versionIdQueryName string) BucketKeyGetter {
	return func(c *gin.Context) (string, string, *string) {
		if versionId, ok := c.GetQuery(versionIdQueryName); ok && versionId != "" {
			return c.Param(bucketParamName), c.Param(keyParamName), &versionId
		}

		return c.Param(bucketParamName), c.Param(keyParamName), nil
	}
}

// WithStaticBucket will retrieve the key from [gin.Context.Param] while the bucket is static.
func WithStaticBucket(bucket, keyParamName string) BucketKeyGetter {
	return func(c *gin.Context) (string, string, *string) {
		return bucket, c.Param(keyParamName), nil
	}
}

// Skipper tells the proxy to skip processing a request.
//
// If the function returns true, the proxy will not act on the request further. Additionally, the function should have
// aborted the request with a meaningful status such as 404.
type Skipper func(c *gin.Context) bool

// ForOnlyMethods can be used to limit the proxy to only certain HTTP methods such as GET and HEAD.
//
// All other methods will get a 404 regardless of whether resource exists.
func ForOnlyMethods(methods ...string) Option {
	m := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		m[method] = struct{}{}
	}

	return func(cfg *Config) {
		cfg.Skip = func(c *gin.Context) bool {
			if _, ok := m[c.Request.Method]; ok {
				return true
			}

			c.AbortWithStatus(http.StatusNotFound)
			return false
		}
	}
}

func statusCode(err error) int {
	var re *awstransporthttp.ResponseError
	if errors.As(err, &re) {
		return re.HTTPStatusCode()
	}

	return 0
}

func handleErr(c *gin.Context, err error) {
	var (
		re   *awstransporthttp.ResponseError
		code int
	)

	if errors.As(err, &re) {
		code = re.HTTPStatusCode()
	}

	switch code {
	case http.StatusNotFound, http.StatusNotModified, http.StatusPreconditionFailed:
		c.AbortWithStatus(code)
	case 0:
		code = http.StatusInternalServerError
		fallthrough
	default:
		_ = c.AbortWithError(code, err)
	}
}
