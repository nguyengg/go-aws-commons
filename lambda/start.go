package lambda

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nguyengg/go-aws-commons/metrics"
	"github.com/rs/zerolog"
)

// StartHandlerFunc is a wrapper around [lambda.StartHandlerFunc] that adds sensible defaults.
//
// What is available out of the box with this wrapper:
//
//  1. SetUpGlobalLogger is used to prefix the AWS Request Id to every log message.
//  2. A metrics.Metrics instance is attached to the context with a few default metrics logged at the end of the request
//     such as start and end time, [metrics.Metrics.Panicked] or [metrics.Metrics.Fault] if the underlying handler
//     panics or returns an error. The wrapper will not attempt to recover so that the stack trace can propagate. The
//     metrics.Metrics instance can be retrieved via metrics.Ctx.
func StartHandlerFunc[TIn any, TOut any](handler func(context.Context, TIn) (TOut, error), options ...lambda.Option) {
	lambda.StartHandlerFunc(func(ctx context.Context, in TIn) (TOut, error) {
		m := metrics.New()
		ctx = metrics.WithContext(ctx, m)
		if lc, ok := lambdacontext.FromContext(ctx); ok {
			m.SetProperty("awsRequestID", lc.AwsRequestID)
		}

		ctx, l := LoggerWithContext(ctx)
		defer SetUpGlobalLogger(ctx)()

		panicked := true
		defer func() {
			if panicked {
				m.Panicked()
			}

			m.Log(l)
		}()

		v, err := handler(ctx, in)
		if err != nil {
			m.Faulted()
		}

		panicked = false
		return v, err
	}, options...)
}

// Start is a variant of StartHandlerFunc for handlers that don't have any explicit returned value.
//
// See StartHandlerFunc for an in-depth explanation on what are available.
func Start[TIn any](handler func(context.Context, TIn) error, options ...lambda.Option) {
	lambda.StartWithOptions(func(ctx context.Context, in TIn) error {
		m := metrics.New()
		ctx = metrics.WithContext(ctx, m)
		if lc, ok := lambdacontext.FromContext(ctx); ok {
			m.SetProperty("awsRequestID", lc.AwsRequestID)
		}

		ctx, l := LoggerWithContext(ctx)
		defer SetUpGlobalLogger(ctx)()

		panicked := true
		defer func() {
			if panicked {
				m.Panicked()
			}

			m.Log(l)
		}()

		err := handler(ctx, in)
		if err != nil {
			m.Faulted()
		}

		panicked = false
		return err
	}, options...)
}

// StartCloudWatchEventHandler logs the CloudWatch event (without Detail attribute) as `event` JSON property.
func StartCloudWatchEventHandler(handler func(context.Context, events.CloudWatchEvent) error) {
	Start(func(ctx context.Context, event events.CloudWatchEvent) error {
		// don't log the detail which is json.RawMessage type.
		sansDetail := event
		sansDetail.Detail = nil

		metrics.Ctx(ctx).SetJSONProperty("event", sansDetail)
		return handler(ctx, event)
	})
}

// StartDynamoDBEventHandler logs the number of records as `recordCount` counter.
func StartDynamoDBEventHandler(handler func(context.Context, events.DynamoDBEvent) error) {
	Start(func(ctx context.Context, event events.DynamoDBEvent) error {
		metrics.Ctx(ctx).AddCount("recordCount", int64(len(event.Records)))
		return handler(ctx, event)
	})
}

// StartDynamoDBEventHandleFunc logs the number of records and the number of batch item failure as `recordCount` and
// `batchItemFailureCount` counters respectively.
func StartDynamoDBEventHandleFunc(handler func(context.Context, events.DynamoDBEvent) (events.DynamoDBEventResponse, error)) {
	StartHandlerFunc(func(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
		m := metrics.Ctx(ctx).AddCount("recordCount", int64(len(event.Records)))
		res, err := handler(ctx, event)
		m.AddCount("batchItemFailureCount", int64(len(res.BatchItemFailures)))
		return res, err
	})
}

// StartS3EventHandler logs the number of records as `recordCount` counter.
func StartS3EventHandler(handler func(context.Context, events.S3Event) error) {
	Start(func(ctx context.Context, event events.S3Event) error {
		metrics.Ctx(ctx).AddCount("recordCount", int64(len(event.Records)))
		return handler(ctx, event)
	})
}

// StartSNSEventHandler logs the number of records as `recordCount` counter.
func StartSNSEventHandler(handler func(context.Context, events.SNSEvent) error) {
	Start(func(ctx context.Context, event events.SNSEvent) error {
		metrics.Ctx(ctx).AddCount("recordCount", int64(len(event.Records)))
		return handler(ctx, event)
	})
}

// StartSQSEventHandler logs the number of records as `recordCount` counter.
func StartSQSEventHandler(handler func(context.Context, events.SQSEvent) error) {
	Start(func(ctx context.Context, event events.SQSEvent) error {
		metrics.Ctx(ctx).AddCount("recordCount", int64(len(event.Records)))
		return handler(ctx, event)
	})
}

// StartSQSEventHandlerFunc logs the number of records and the number of batch item failure as `recordCount` and
// `batchItemFailureCount` counters respectively.
func StartSQSEventHandlerFunc(handler func(context.Context, events.SQSEvent) (events.SQSEventResponse, error)) {
	StartHandlerFunc(func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
		m := metrics.Ctx(ctx).AddCount("recordCount", int64(len(event.Records)))
		res, err := handler(ctx, event)
		m.AddCount("batchItemFailureCount", int64(len(res.BatchItemFailures)))
		return res, err
	})
}

// LoggerWithContext returns a valid zerolog.Logger for use.
//
// Because zerolog.Ctx may return a disabled (no-op) logger, it's difficult to determine if user is intentionally
// disabling logging via context, or if the zerolog.DefaultContextLogger has not been set up. As a result, this method
// may create a new logger if it can determine that one should be created, and the logger will be attached to the
// new returned context in that case. If zerolog.DefaultContextLogger is not nil then the returned value from
// zerolog.Ctx is always used.
//
// Furthermore, if a Lambda context is available from lambdacontext.FromContext, the zerolog.Context of the returned
// logger is updated with a string awsRequestID.
func LoggerWithContext(ctx context.Context) (_ context.Context, l *zerolog.Logger) {
	if l = zerolog.Ctx(ctx); l.GetLevel() == zerolog.Disabled && zerolog.DefaultContextLogger == nil {
		newLogger := zerolog.New(os.Stderr)
		l = &newLogger
		ctx = l.WithContext(ctx)
	}

	if lc, ok := lambdacontext.FromContext(ctx); ok {
		l.UpdateContext(func(c zerolog.Context) zerolog.Context {
			c.Str("awsRequestID", lc.AwsRequestID)
			return c
		})
	}

	return ctx, l
}
