package lambda

import (
	awslambda "github.com/aws/aws-lambda-go/lambda"
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
func StartHandlerFunc[TIn any, TOut any, H awslambda.HandlerFunc[TIn, TOut]](handler H, options ...awslambda.Option) {
	NewHandlerFunc(handler, options...).Start()
}

// Start is a variant of StartHandlerFunc for handlers that don't have any explicit returned value.
//
// See StartHandlerFunc for details.
func Start[TIn any, H ConsumerFunc[TIn]](handler H, options ...awslambda.Option) {
	NewConsumerFunc(handler, options...).Start()
}
