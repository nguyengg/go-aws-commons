package lambda

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nguyengg/go-aws-commons/metrics"
)

// StartHandlerFunc is a wrapper around [lambda.StartHandlerFunc] that adds sensible logging and metrics out of the box.
func StartHandlerFunc[TIn any, TOut any](handler func(context.Context, TIn) (TOut, error), options ...lambda.Option) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{ReplaceAttr: metrics.ReplaceAttr()})))

	lambda.StartHandlerFunc(func(ctx context.Context, in TIn) (out TOut, err error) {
		ctx, m := setUp(ctx)

		defer func() {
			switch r := recover(); {
			case r != nil:
				m.Panicked()
				m.Any("error", r)
				slog.LogAttrs(ctx, slog.LevelError, "invocation panicked", m.Attrs()...)
				panic(r)

			case err != nil:
				m.Faulted()
				m.Any("error", err)
				slog.LogAttrs(ctx, slog.LevelError, "invocation error", m.Attrs()...)

			default:
				slog.LogAttrs(ctx, slog.LevelInfo, "invocation done", m.Attrs()...)
			}
		}()

		out, err = handler(ctx, in)
		return out, err
	}, options...)
}

// Start is a variant of StartHandlerFunc for handlers that don't have any explicit returned value.
//
// See StartHandlerFunc for an in-depth explanation on what are available.
func Start[TIn any](handler func(context.Context, TIn) error, options ...lambda.Option) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{ReplaceAttr: metrics.ReplaceAttr()})))

	lambda.StartWithOptions(func(ctx context.Context, in TIn) (err error) {
		ctx, m := setUp(ctx)

		defer func() {
			switch r := recover(); {
			case r != nil:
				m.Panicked()
				m.Any("error", r)
				slog.LogAttrs(ctx, slog.LevelError, "invocation panicked", m.Attrs()...)
				panic(r)

			case err != nil:
				m.Faulted()
				m.Any("error", err)
				slog.LogAttrs(ctx, slog.LevelError, "invocation error", m.Attrs()...)

			default:
				slog.LogAttrs(ctx, slog.LevelInfo, "invocation done", m.Attrs()...)
			}
		}()

		err = handler(ctx, in)
		return err
	}, options...)
}

// setUp applies sensible default settings to log.Default and slog.Default instances while also attaching a Metrics
// instance to the returned context.
func setUp(ctx context.Context) (context.Context, *metrics.Metrics) {
	if IsDebug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		log.SetFlags(DebugLogFlags)
	} else {
		log.SetFlags(DefaultLogFlags)
	}

	m := metrics.New()

	if lc, ok := lambdacontext.FromContext(ctx); ok {
		log.SetPrefix(lc.AwsRequestID + " ")
		slog.SetDefault(slog.With("awsRequestId", lc.AwsRequestID))
		m.String("awsRequestId", lc.AwsRequestID)
	}

	return metrics.WithContext(ctx, m), m
}
