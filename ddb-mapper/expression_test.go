package ddb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/stretchr/testify/assert"
)

func TestExpressionError(t *testing.T) {
	// if AWS changes the error message, I want to know about it to help the user understand that they must provide
	// an update expression.
	_, err := expression.NewBuilder().WithUpdate(expression.UpdateBuilder{}).Build()
	assert.ErrorContains(t, err, "unset parameter: UpdateBuilder")
}
