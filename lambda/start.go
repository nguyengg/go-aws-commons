package lambda

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nguyengg/go-aws-commons/metrics"

	"github.com/rotisserie/eris"
)

// StartHandlerFunc is a wrapper around [lambda.StartHandlerFunc] that adds sensible logging and metrics out of the box.
//
// A metrics.Metrics instance will be made available via context by way of metrics.Get. The "fault" and "panicked"
// counters will be populated accordingly: if the handler returns a non-nil error, "fault" will be set to 1, and "error"
// property to the error; if the handler panics, "panicked" will be set to 1, "error" to the recovered error, and
// "stack" to the stack trace.
//
// See SetUpLogDefault and SetUpSlogDefault for how the default loggers are configured. Additionally, for each Lambda
// invocation, the AWS request Id will also be attached to log.Default as the message prefix, slog.Default as the
// attribute "awsRequestId". To disable this feature, specify either HandlerFuncOptions.NoSetUpLogging or
// ConsumerFuncOptions.NoSetUpLogging.
//
// If you need to customise the metrics.Metrics instance, use NewHandlerFunc.
func StartHandlerFunc[TIn any, TOut any, H lambda.HandlerFunc[TIn, TOut]](handler H, options ...lambda.Option) {
	NewHandlerFunc(handler, options...).Start()
}

// Start is a variant of StartHandlerFunc for handlers that don't have any explicit returned value.
//
// See StartHandlerFunc for details.
func Start[TIn any, H ConsumerFunc[TIn]](handler H, options ...lambda.Option) {
	NewConsumerFunc(handler, options...).Start()
}

// NewHandlerFunc is a more customisable variant of StartHandlerFunc.
//
// You must eventually call HandlerFuncOptions.Start to enter the Lambda loop.
func NewHandlerFunc[TIn any, TOut any, H lambda.HandlerFunc[TIn, TOut]](handler H, options ...lambda.Option) *HandlerFuncOptions[TIn, TOut, H] {
	return &HandlerFuncOptions[TIn, TOut, H]{
		handler: handler,
		options: options,
	}
}

// HandlerFuncOptions is returned by NewHandlerFunc to allow further customisation.
type HandlerFuncOptions[TIn any, TOut any, H lambda.HandlerFunc[TIn, TOut]] struct {
	// NoSetUpLogging will disable logging features.
	NoSetUpLogging bool

	handler        H
	options        []lambda.Option
	metricsOptions []func(m *metrics.Metrics)
}

// WithMetricsOptions replaces the customisations for metrics.New.
func (h *HandlerFuncOptions[TIn, TOut, H]) WithMetricsOptions(optFns ...func(m *metrics.Metrics)) *HandlerFuncOptions[TIn, TOut, H] {
	h.metricsOptions = optFns
	return h
}

// ConsumerFunc is a variant of lambda.HandlerFunc that has no output (except for error).
type ConsumerFunc[TIn any] interface {
	func(context.Context, TIn) error
}

// NewConsumerFunc is a more customisable variant of Start.
//
// You must eventually call ConsumerFuncOptions.Start to enter the Lambda loop.
func NewConsumerFunc[TIn any, H ConsumerFunc[TIn]](handler H, options ...lambda.Option) *ConsumerFuncOptions[TIn, H] {
	return &ConsumerFuncOptions[TIn, H]{
		handler: handler,
		options: options,
	}
}

// WithMetricsOptions replaces the customisations for metrics.New.
func (h *ConsumerFuncOptions[TIn, H]) WithMetricsOptions(optFns ...func(m *metrics.Metrics)) *ConsumerFuncOptions[TIn, H] {
	h.metricsOptions = optFns
	return h
}

// ConsumerFuncOptions is returned by NewConsumerFunc to allow further customisation.
type ConsumerFuncOptions[TIn any, H ConsumerFunc[TIn]] struct {
	// NoSetUpLogging will disable logging features.
	NoSetUpLogging bool

	handler        H
	options        []lambda.Option
	metricsOptions []func(m *metrics.Metrics)
}

// Start is a wrapper around [lambda.StartHandlerFunc] that adds sensible logging and metrics out of the box.
//
// See package-level Start for details.
func (h *HandlerFuncOptions[TIn, TOut, H]) Start() {
	if !h.NoSetUpLogging {
		SetUpLogDefault()
		SetUpSlogDefault()
	}

	lambda.StartHandlerFunc(func(ctx context.Context, in TIn) (out TOut, err error) {
		ctx, m := metrics.NewWithContext(ctx, h.metricsOptions...)

		if lc, ok := lambdacontext.FromContext(ctx); ok {
			if !h.NoSetUpLogging {
				log.SetPrefix(lc.AwsRequestID + " ")
				slog.SetDefault(slog.With("awsRequestId", lc.AwsRequestID))
			}

			m.String("awsRequestId", lc.AwsRequestID)
		}

		defer func() {
			switch r := recover(); {
			case r != nil:
				m.Panicked()

				switch v := r.(type) {
				case error:
					m.Error(v)
				default:
					m.Error(eris.Wrapf(fmt.Errorf("%+v", v), "recover non-error %T: %#v", v, v))
				}

			case err != nil:
				m.Error(err)

			}

			_ = m.CloseContext(ctx)
		}()

		out, err = h.handler(ctx, in)
		return out, err
	}, h.options...)
}

// Start is a variant of StartHandlerFunc for handlers that don't have any explicit returned value.
//
// See package-level StartHandlerFunc for details.
func (h *ConsumerFuncOptions[TIn, H]) Start() {
	if !h.NoSetUpLogging {
		SetUpLogDefault()
		SetUpSlogDefault()
	}

	lambda.StartWithOptions(func(ctx context.Context, in TIn) (err error) {
		ctx, m := metrics.NewWithContext(ctx, h.metricsOptions...)

		if lc, ok := lambdacontext.FromContext(ctx); ok {
			if !h.NoSetUpLogging {
				log.SetPrefix(lc.AwsRequestID + " ")
				slog.SetDefault(slog.With("awsRequestId", lc.AwsRequestID))
			}

			m.String("awsRequestId", lc.AwsRequestID)
		}

		defer func() {
			switch r := recover(); {
			case r != nil:
				m.Panicked()

				switch v := r.(type) {
				case error:
					m.Error(v)
				default:
					m.Error(eris.Wrapf(fmt.Errorf("%+v", v), "recover non-error %T: %#v", v, v))
				}

			case err != nil:
				m.Error(err)

			}

			_ = m.CloseContext(ctx)
		}()

		err = h.handler(ctx, in)
		return err
	}, h.options...)
}
