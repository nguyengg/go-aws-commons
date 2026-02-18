package lambda

import (
	"context"
	"log"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nguyengg/go-aws-commons/metrics"
)

// StartHandlerFunc is a wrapper around [lambda.StartHandlerFunc] that adds sensible logging and metrics out of the box.
//
// If you need to customise the metrics.Metrics instance, use NewHandlerFunc.
func StartHandlerFunc[TIn any, TOut any, H lambda.HandlerFunc[TIn, TOut]](handler H, options ...lambda.Option) {
	NewHandlerFunc(handler, options...).Start()
}

// Start is a variant of StartHandlerFunc for handlers that don't have any explicit returned value.
//
// See StartHandlerFunc for an in-depth explanation on what are available.
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
	handler        H
	options        []lambda.Option
	metricsOptions []func(m *metrics.Metrics)
}

func (h *HandlerFuncOptions[TIn, TOut, H]) Start() {
	if IsDebug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		log.SetFlags(DebugLogFlags)
	} else {
		log.SetFlags(DefaultLogFlags)
	}

	lambda.StartHandlerFunc(func(ctx context.Context, in TIn) (out TOut, err error) {
		ctx, m := metrics.NewWithContext(ctx, h.metricsOptions...)

		if lc, ok := lambdacontext.FromContext(ctx); ok {
			log.SetPrefix(lc.AwsRequestID + " ")
			slog.SetDefault(slog.With("awsRequestId", lc.AwsRequestID))
			m.String("awsRequestId", lc.AwsRequestID)
		}

		defer func() {
			switch r := recover(); {
			case r != nil:
				m.Panicked()
				m.Any("error", r)
				_ = m.CloseContext(ctx)
				panic(r)

			case err != nil:
				m.Faulted()
				m.Any("error", err)
				_ = m.CloseContext(ctx)

			default:
				_ = m.CloseContext(ctx)
			}
		}()

		out, err = h.handler(ctx, in)
		return out, err
	}, h.options...)
}

func (h *ConsumerFuncOptions[TIn, H]) Start() {
	if IsDebug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		log.SetFlags(DebugLogFlags)
	} else {
		log.SetFlags(DefaultLogFlags)
	}

	lambda.StartWithOptions(func(ctx context.Context, in TIn) (err error) {
		ctx, m := metrics.NewWithContext(ctx, h.metricsOptions...)

		if lc, ok := lambdacontext.FromContext(ctx); ok {
			log.SetPrefix(lc.AwsRequestID + " ")
			slog.SetDefault(slog.With("awsRequestId", lc.AwsRequestID))
			m.String("awsRequestId", lc.AwsRequestID)
		}

		defer func() {
			switch r := recover(); {
			case r != nil:
				m.Panicked()
				m.Any("error", r)
				_ = m.CloseContext(ctx)
				panic(r)

			case err != nil:
				m.Faulted()
				m.Any("error", err)
				_ = m.CloseContext(ctx)

			default:
				_ = m.CloseContext(ctx)
			}
		}()

		err = h.handler(ctx, in)
		return err
	}, h.options...)
}
