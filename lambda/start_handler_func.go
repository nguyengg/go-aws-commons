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
				slog.SetDefault(slog.With(slog.String("awsRequestId", lc.AwsRequestID)))
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

			case err != nil && !m.HasError():
				m.Error(err)

			}

			_ = m.CloseContext(ctx)
		}()

		out, err = h.handler(ctx, in)
		return out, err
	}, h.options...)
}

// WithMetricsOptions replaces the customisations for metrics.New.
func (h *HandlerFuncOptions[TIn, TOut, H]) WithMetricsOptions(optFns ...func(m *metrics.Metrics)) *HandlerFuncOptions[TIn, TOut, H] {
	h.metricsOptions = optFns
	return h
}
