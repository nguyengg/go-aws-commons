package lambda

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"log"
)

// RecommendedLogFlag is the flag passed to log.SetFlags by SetUpLogger.
const RecommendedLogFlag = log.Ldate | log.Lmicroseconds | log.LUTC | log.Lmsgprefix | log.Lshortfile

// SetUpGlobalLogger applies sensible default settings to  log.Default.
//
// Specifically, [log.SetFlags] is called with RecommendedLogFlag, and if [lambdacontext.LambdaContext.AwsRequestId] is
// available then it is set as the log prefix with [log.SetPrefix].
//
// A function is returned that should be deferred upon to reset the log flags and prefix back to the original values.
// Use SetUpLogger if you wish to modify a specific log.Logger.
//
// Usage
//
//	// this should be the first line in your AWS Lambda handler. many Start methods in this package will do this
//	// for you by default.
//	// notice the double ()() to make sure SetUpGlobalLogger executes some function first, then its returned
//	// function is deferred.
//	defer logsupport.SetUpGlobalLogger()()
func SetUpGlobalLogger(ctx context.Context) func() {
	return SetUpLogger(ctx, log.Default())
}

// SetUpLogger is a variant of SetUpGlobalLogger that targets a specific log.Logger.
func SetUpLogger(ctx context.Context, logger *log.Logger) func() {
	flags := logger.Flags()
	prefix := logger.Prefix()

	logger.SetFlags(RecommendedLogFlag)

	if lc, ok := lambdacontext.FromContext(ctx); ok {
		logger.SetPrefix(lc.AwsRequestID + " ")
	}

	return func() {
		logger.SetFlags(flags)
		logger.SetPrefix(prefix)
	}
}
