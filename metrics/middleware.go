package metrics

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmw "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"
	smithymw "github.com/aws/smithy-go/middleware"
)

// WithClientSideMetrics adds a ClientSideMetricsMiddleware to the config that is being created.
//
// Usage:
//
//	cfg, err := config.LoadDefaultConfig(context.TODO(), middleware.WithClientSideMetrics())
//
// See ClientSideMetricsMiddleware for what kind of metrics are populated.
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
//
//	// alternatively
//	configcache.Get(ctx, metrics.AddClientSideMetrics)
//
// See ClientSideMetricsMiddleware for what kind of metrics are populated.
func AddClientSideMetrics(cfg *aws.Config) {
	cfg.APIOptions = append(cfg.APIOptions, ClientSideMetricsMiddleware())
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
// A Metrics instance must be available from context by the time the middleware receives a response. That instance's
// counters and timings metrics will be populated with the metrics from the AWS service calls. For example, if one S3
// GetObject one DynamoDB Query call were made using the same Metrics instance, it will be populated with counters like
// this:
//
//	"counters": {
//	    "S3.GetObject.ClientFault": 0,
//	    "S3.GetObject.ServerFault": 0,
//	    "S3.GetObject.UnknownFault": 0,
//	    "DynamoDB.Query.ClientFault": 0,
//	    "DynamoDB.Query.ServerFault": 0,
//	    "DynamoDB.Query.UnknownFault": 0,
//	},
//	"timings": {
//	    "S3.GetObject": {
//	        "sum": "64.680ms",
//	        "min": "64.680ms",
//	        "max": "64.680ms",
//	        "n": 1,
//	        "avg": "64.680ms"
//	    },
//	    "DynamoDB.Query": {
//	        "sum": "74.255ms",
//	        "min": "74.255ms",
//	        "max": "74.255ms",
//	        "n": 1,
//	        "avg": "74.255ms"
//	    }
//	}
//
// Note that the middleware does not do any logging on its own; it only populates the Metrics instance attached to
// the context passed into the AWS calls.
func ClientSideMetricsMiddleware(options ...Option) func(stack *smithymw.Stack) error {
	middleware := &clientSideMetricsMiddleware{}
	for _, fn := range options {
		fn(middleware)
	}

	return func(stack *smithymw.Stack) error {
		return stack.Deserialize.Add(middleware, smithymw.After)
	}
}

type clientSideMetricsMiddleware struct {
}

func (c clientSideMetricsMiddleware) ID() string {
	return "ClientSideLatencyMetrics"
}

func (c clientSideMetricsMiddleware) HandleDeserialize(ctx context.Context, in smithymw.DeserializeInput, next smithymw.DeserializeHandler) (out smithymw.DeserializeOutput, metadata smithymw.Metadata, err error) {
	start := time.Now()

	out, metadata, err = next.HandleDeserialize(ctx, in)

	end := time.Now()
	if t, ok := awsmw.GetResponseAt(metadata); ok {
		end = t
	}

	m, ok := TryGet(ctx)
	if !ok {
		slog.LogAttrs(ctx, slog.LevelWarn, "no metrics available from context")
		return
	}

	// DynamoDB GetItem => DynamoDB.GetItem
	// log filter can use { $.['DynamoDB.GetItem.ServerFault'] = * }
	// see https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html#matching-terms-json-log-events
	serviceId := awsmw.GetServiceID(ctx)
	operationName := awsmw.GetOperationName(ctx)
	key := serviceId + "." + operationName
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
			m.AddCounter(key+".ClientFault", 1, key+".ServerFault")
		case smithy.FaultServer:
			m.AddCounter(key+".ServerFault", 1, key+".ClientFault")
		case smithy.FaultUnknown:
			fallthrough
		default:
			m.AddCounter(key+".UnknownFault", 1, key+".ClientFault", key+".ServerFault")
		}
	} else {
		m.SetCounter(key+".ClientFault", 0, key+".ServerFault")
	}

	return
}

// Option allows customization of the ClientSideMetricsMiddleware.
//
// At the moment, there are no customizations available yet.
type Option func(*clientSideMetricsMiddleware)
