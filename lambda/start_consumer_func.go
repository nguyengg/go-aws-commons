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

// ConsumerFuncOptions is returned by NewConsumerFunc to allow further customisation.
type ConsumerFuncOptions[TIn any, H ConsumerFunc[TIn]] struct {
	// NoSetUpLogging will disable logging features.
	NoSetUpLogging bool

	handler        H
	options        []lambda.Option
	metricsOptions []func(m *metrics.Metrics)
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

		err = h.handler(ctx, in)
		return err
	}, h.options...)
}

// WithMetricsOptions replaces the customisations for metrics.New.
func (h *ConsumerFuncOptions[TIn, H]) WithMetricsOptions(optFns ...func(m *metrics.Metrics)) *ConsumerFuncOptions[TIn, H] {
	h.metricsOptions = optFns
	return h
}
