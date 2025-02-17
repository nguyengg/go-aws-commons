# Gin adapter for Function URL

A Gin adapter for API Gateway V1 and V2 are already available from https://github.com/awslabs/aws-lambda-go-api-proxy.
This module provides an adapter specifically for Function URL events with both BUFFERED (which, technically, is no
different from API Gateway V2/HTTP events) and RESPONSE_STREAM mode. The adapter for RESPONSE_STREAM actually is an
entirely custom and experimental runtime built from specifications available at
https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html#runtimes-custom-response-streaming.
