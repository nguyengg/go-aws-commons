package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmw "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"
	smithymw "github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/rs/zerolog"
)

// WithClientSideMetrics adds a ClientSideMetricsMiddleware to the config that is being created.
//
// Usage:
//
//	cfg, err := config.LoadDefaultConfig(context.TODO(), metrics.WithClientSideMetrics())
func WithClientSideMetrics(options ...Option) func(*config.LoadOptions) error {
	return func(cfg *config.LoadOptions) error {
		cfg.APIOptions = append(cfg.APIOptions, ClientSideMetricsMiddleware(options...))
		return nil
	}
}

// AddClientSideMetrics adds a ClientSideMetricsMiddleware to the config that has been created and is passed here.
//
// Usage:
//
//	cfg, _ := config.LoadDefaultConfig(ctx)
//	metrics.AddClientSideMetrics(cfg)
func AddClientSideMetrics(cfg *aws.Config, options ...Option) {
	cfg.APIOptions = append(cfg.APIOptions, ClientSideMetricsMiddleware(options...))
}

// ClientSideMetricsMiddleware creates a new middleware to add client-side latency metrics about the requests.
//
// Usage:
//
//	cfg, _ := config.LoadDefaultConfig(ctx)
//	cfg.APIOptions = append(cfg.APIOptions, metrics.ClientSideMetricsMiddleware())
//
//	// alternatively
//	metrics.AddClientSideMetrics(cfg)
//
// A metrics.Metrics instance must be available from context by the time the middleware receives a response. By default,
// zerolog.Ctx is used to retrieve a zerolog.Logger instance that logs the metrics instance. This can be customised via
// WithLogger.
func ClientSideMetricsMiddleware(options ...Option) func(stack *smithymw.Stack) error {
	middleware := &clientSideMetricsMiddleware{
		logFn: zerolog.Ctx,
	}
	for _, fn := range options {
		fn(middleware)
	}

	return func(stack *smithymw.Stack) error {
		return stack.Deserialize.Add(middleware, smithymw.After)
	}
}

// Should implement middleware.DeserializeMiddleware.
type clientSideMetricsMiddleware struct {
	logFn func(context.Context) *zerolog.Logger
}

// Option allows customization of the ClientSideMetricsMiddleware.
//
// At the moment, there are no customizations available yet.
type Option func(*clientSideMetricsMiddleware)

// WithLogger changes the zerolog.Logger instance that is used.
//
// By default, zerolog.Ctx is used to retrieve the logger from context.
func WithLogger(logFn func(context.Context) *zerolog.Logger) Option {
	return func(middleware *clientSideMetricsMiddleware) {
		middleware.logFn = logFn
	}
}

func (c *clientSideMetricsMiddleware) ID() string {
	return "ClientSideLatencyMetrics"
}

func (c *clientSideMetricsMiddleware) HandleDeserialize(ctx context.Context, input smithymw.DeserializeInput, handler smithymw.DeserializeHandler) (smithymw.DeserializeOutput, smithymw.Metadata, error) {
	start := time.Now().UTC()

	output, metadata, err := handler.HandleDeserialize(ctx, input)

	end := time.Now().UTC()
	if t, ok := awsmw.GetResponseAt(metadata); ok {
		end = t
	}

	serviceId := awsmw.GetServiceID(ctx)
	operationName := awsmw.GetOperationName(ctx)

	e := c.logFn(ctx).
		Log().
		Str("service", serviceId).
		Str("operation", operationName).
		Int64(ReservedKeyStartTime, start.UnixNano()/int64(time.Millisecond)).
		Str(ReservedKeyEndTime, end.Format(http.TimeFormat)).
		Str(ReservedKeyTime, FormatDuration(end.Sub(start)))

	counters := zerolog.Dict()

	switch resp := output.RawResponse.(type) {
	case *smithyhttp.Response:
		e.Int("statusCode", resp.StatusCode)

		switch resp.StatusCode / 100 {
		case 2:
			counters.Int("2xx", 1).Int("4xx", 0).Int("5xx", 0)
		case 4:
			counters.Int("2xx", 0).Int("4xx", 1).Int("5xx", 0)
		case 5:
			counters.Int("2xx", 0).Int("4xx", 0).Int("5xx", 1)
		default:
			counters.Int("2xx", 0).Int("4xx", 0).Int("5xx", 0)
		}
	}

	// DynamoDB GetItem => DynamoDB.GetItem
	// log filter can use { $.['DynamoDB.GetItem.Fault'] = * }
	// see https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html#matching-terms-json-log-events
	key := serviceId + "." + operationName

	m := Ctx(ctx)
	m.AddTiming(key, end.Sub(start))

	if err != nil {
		// check whether is server fault or not.
		var ae smithy.APIError
		var errorFault smithy.ErrorFault
		if errors.As(err, &ae) {
			errorFault = ae.ErrorFault()
		}

		switch errorFault {
		case smithy.FaultClient:
			m.AddCount(key+".ClientFault", 1, key+".ServerFault", key+".UnknownFault")
			counters.Int("clientFault", 1).Int("serverFault", 0).Int("unknownFault", 0)
			e.AnErr("clientError", err)
		case smithy.FaultServer:
			m.AddCount(key+".ServerFault", 1, key+".ClientFault", key+".UnknownFault")
			counters.Int("clientFault", 0).Int("serverFault", 1).Int("unknownFault", 0)
			e.AnErr("serverError", err)
		case smithy.FaultUnknown:
			fallthrough
		default:
			m.AddCount(key+".UnknownFault", 1, key+".ClientFault", key+".ServerFault")
			counters.Int("clientFault", 0).Int("serverFault", 0).Int("unknownFault", 1)
			e.AnErr("unknownError", err)
		}
	} else {
		m.AddCount(key+".ClientFault", 0, key+".ServerFault", key+".UnknownFault")
		counters.Int("clientFault", 0).Int("serverFault", 0).Int("unknownFault", 0)
	}

	e.Dict("counters", counters).Msg("")

	return output, metadata, err
}
