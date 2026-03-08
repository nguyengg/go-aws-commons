package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_equivalent(t *testing.T) {
	type Item struct {
		ID string `dynamodbav:"id,hashkey" tablename:"Items"`
	}

	m1, err := NewForType[Item]()
	assert.NoError(t, err, "NewForType[Item]()")
	m2, err := NewForType[*Item]()
	assert.NoError(t, err, "NewForType[*Item]()")
	m3, err := NewForTypeOf(Item{})
	assert.NoError(t, err, "NewForTypeOf(Item{})")
	m4, err := NewForTypeOf(&Item{})
	assert.NoError(t, err, "NewForTypeOf(&Item{})")

	assert.Equal(t, m1, m2)
	assert.Equal(t, m2, m3)
	assert.Equal(t, m3, m4)
}
