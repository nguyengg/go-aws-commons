package lambda

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// SetUpSlogDefault sets the default slog.Default to print JSON messages to os.Stderr, respecting AWS_LAMBDA_LOG_FORMAT
// and AWS_LAMBDA_LOG_LEVEL if given.
//
// Replicates the logic of lambdacontext.NewLogHandler with a few key differences:
//   - "awsRequestId" is used instead of "requestId".
//   - defaults to JSON handler if AWS_LAMBDA_LOG_FORMAT is not given.
//   - if DEBUG environment variable is true-ish, override AWS_LAMBDA_LOG_LEVEL to DEBUG.
func SetUpSlogDefault() {
	opts := &slog.HandlerOptions{ReplaceAttr: lambdacontext.ReplaceAttr}

	if debug, _ := strconv.ParseBool(os.Getenv("DEBUG")); debug {
		opts.Level = slog.LevelDebug
	} else {
		switch os.Getenv("AWS_LAMBDA_LOG_LEVEL") {
		case "DEBUG":
			opts.Level = slog.LevelDebug
		case "INFO":
			opts.Level = slog.LevelInfo
		case "WARN":
			opts.Level = slog.LevelWarn
		case "ERROR":
			opts.Level = slog.LevelError
		}
	}

	// this is where we deviate from lambdacontext.NewLogHandler: if AWS_LAMBDA_FORMAT is not given, default to JSON.
	// if AWS_LAMBDA_FORMAT is given then follow what lambdacontext.NewLogHandler does.
	var handler slog.Handler
	switch format, ok := os.LookupEnv("AWS_LAMBDA_FORMAT"); {
	case format == "JSON":
		fallthrough
	case !ok:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(&slogHandler{handler}))
}

type slogHandler struct {
	slog.Handler
}

func (s slogHandler) Handle(ctx context.Context, record slog.Record) error {
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		record.AddAttrs(slog.String("awsRequestId", lc.AwsRequestID))
	}

	return s.Handler.Handle(ctx, record)
}

func (s slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{s.WithAttrs(attrs)}
}

func (s slogHandler) WithGroup(name string) slog.Handler {
	return &slogHandler{s.WithGroup(name)}
}
