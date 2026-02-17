package lambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nguyengg/go-aws-commons/metrics"
)

// StartSQSMessageHandler starts the Lambda loop for handling SQS messages in a batch.
//
// See https://docs.aws.amazon.com/prescriptive-guidance/latest/lambda-event-filtering-partial-batch-responses-for-sqs/benefits-partial-batch-responses.html
// for why partial batch response is preferable over failing the entire batch.
//
// The handler works on one message at a time and sequentially. If the handler returns a non-nil error, the wrapper will
// automatically add an events.SQSBatchItemFailure so that only failed messages get retried later.
func StartSQSMessageHandler(handler func(context.Context, events.SQSMessage) error, options ...lambda.Option) {
	StartHandlerFunc(func(ctx context.Context, req events.SQSEvent) (events.SQSEventResponse, error) {
		m := metrics.Get(ctx)

		res := events.SQSEventResponse{
			BatchItemFailures: make([]events.SQSBatchItemFailure, 0),
		}

		m.AddCounter("recordCount", int64(len(req.Records)), "failureCount")

		for _, record := range req.Records {
			if err := handler(ctx, record); err != nil {
				m.AddCounter("failureCount", 1)
				res.BatchItemFailures = append(res.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
			}
		}

		// very important that nil error is returned here.
		return res, nil
	}, options...)
}
