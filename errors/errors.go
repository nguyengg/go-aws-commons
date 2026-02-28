// Package errors provides convenient methods to extract status code and other metadata from AWS errors.
package errors

import (
	"errors"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/smithy-go"
)

// StatusCode extracts the [awshttp.ResponseError.HTTPStatusCode] from the given error.
//
// Generally you should prefer using [errors.As] on the particular error you're trying to catch (see
// https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/handle-errors.html), but if it's insufficient or doesn't
// work (e.g. s3.IsNoSuchKey) then use this.
//
// Return 0 if err is not an [awshttp.ResponseError].
func StatusCode(err error) int {
	var re *awshttp.ResponseError
	if errors.As(err, &re) {
		return re.HTTPStatusCode()
	}

	return 0
}

// Extract uses [errors.As] to check if the given error is [http.ResponseError], [smithy.APIError] and/or
// [smithy.OperationError].
//
// The returned statusCode comes from [http.ResponseError]; service and operation from smithy.OperationError; the rest
// smithy.APIError.
//
// Generally you should prefer using [errors.As] on the particular error you're trying to catch (see
// https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/handle-errors.htm), but if it's insufficient or doesn't
// work (e.g. s3.IsNoSuchKey) then use this.
//
// If you only need status code, use StatusCode.
func Extract(err error) (statusCode int, service, operation, code, message string, fault smithy.ErrorFault) {
	var re *awshttp.ResponseError
	if errors.As(err, &re) {
		statusCode = re.HTTPStatusCode()
	}

	var ae smithy.APIError
	if errors.As(err, &ae) {
		code = ae.ErrorCode()
		message = ae.ErrorMessage()
		fault = ae.ErrorFault()
	}

	var oe *smithy.OperationError
	if errors.As(err, &oe) {
		service = oe.Service()
		operation = oe.Operation()
	}

	return
}
