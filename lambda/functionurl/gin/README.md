# Gin adapter for Function URL

A Gin adapter for API Gateway V1 and V2 are already available from https://github.com/awslabs/aws-lambda-go-api-proxy.
This module provides an adapter specifically for Function URL events with both BUFFERED (which, technically, is no
different from API Gateway V2/HTTP events) and RESPONSE_STREAM mode which uses
https://github.com/aws/aws-lambda-go/tree/main/lambdaurl.
