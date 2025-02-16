package lambda

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// IsDebug is true if the "DEBUG" environment have value "1" or "true".
//
// The value of IsDebug is set at startup by way of init(). While many things in the lambda package use this value,
// nothing will modify it. If you want to use a different environment variable or a different way to toggle DEBUG
// behaviour, modify this value directly.
var IsDebug bool

func init() {
	switch os.Getenv("DEBUG") {
	case "1", "true":
		IsDebug = true
	}
}

const (
	// DebugLogFlags is the flag passed to log.SetFlags by SetUpLogger if IsDebug is true.
	DebugLogFlags = log.Ldate | log.Lmicroseconds | log.LUTC | log.Llongfile | log.Lmsgprefix

	// DefaultLogFlags is the flag passed to log.SetFlags by SetUpLogger if IsDebug is false.
	DefaultLogFlags = DebugLogFlags | log.Lshortfile
)

// SetUpGlobalLogger applies sensible default settings to log.Default instance.
//
// Specifically, [log.SetFlags] is called with DefaultLogFlags, and if [lambdacontext.LambdaContext.AwsRequestId] is
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

	if IsDebug {
		logger.SetFlags(DebugLogFlags)
	} else {
		logger.SetFlags(DefaultLogFlags)
	}

	if lc, ok := lambdacontext.FromContext(ctx); ok {
		logger.SetPrefix(lc.AwsRequestID + " ")
	}

	return func() {
		logger.SetFlags(flags)
		logger.SetPrefix(prefix)
	}
}
