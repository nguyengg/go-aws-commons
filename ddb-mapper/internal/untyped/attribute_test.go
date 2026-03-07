package untyped

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttribute_GetAndSet(t *testing.T) {
	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	item := Item{}

	// Get with struct type
	v, err := m.HashKey.Get(item)
	require.NoError(t, err)
	assert.Equal(t, "", v)
	require.NoError(t, m.HashKey.Set(&item, "my-id")) // Set must be on pointer

	// Get with struct pointer type
	v, err = m.HashKey.Get(&item)
	require.NoError(t, err)
	assert.Equal(t, "my-id", v)

	// Get with pointer to struct pointer type
	pointerToItem := &Item{"new-id"}
	v, err = m.HashKey.Get(&pointerToItem)
	require.NoError(t, err)
	assert.Equal(t, "new-id", v)
	require.NoError(t, m.HashKey.Set(&pointerToItem, "another-id")) // Set on pointer to struct pointer

	// Get on struct pointer type
	v, err = m.HashKey.Get(pointerToItem)
	require.NoError(t, err)
	assert.Equal(t, "another-id", v)
}
